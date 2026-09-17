package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/embed"
	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/qdrant"
	"github.com/google/uuid"
)

type StoreTarget string

const (
	TargetSQLite StoreTarget = "sqlite"
	TargetQdrant StoreTarget = "qdrant"
	TargetBoth   StoreTarget = "both"
)

func ValidKind(k string) bool {
	switch k {
	case "general_knowledge", "expertise", "resolved_issues", "task_guide":
		return true
	}
	return false
}

func ValidTarget(t string) bool {
	switch StoreTarget(t) {
	case TargetSQLite, TargetQdrant, TargetBoth:
		return true
	}
	return false
}

type Item struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Tags      []string       `json:"tags"`
	Metadata  map[string]any `json:"metadata"`
	InSQLite  bool           `json:"in_sqlite"`
	InQdrant  bool           `json:"in_qdrant"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Score     *float64       `json:"score,omitempty"`
}

type Service struct {
	DB     *sql.DB
	Qdrant *qdrant.Client
	Embed  embed.Embedder
}

func snippet(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func marshalTags(tags []string) (string, error) {
	if tags == nil {
		tags = []string{}
	}
	b, err := json.Marshal(tags)
	return string(b), err
}

type scannable interface{ Scan(dest ...any) error }

func parseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func scanItem(row scannable) (*Item, error) {
	var it Item
	var tagsRaw, metaRaw string
	var inSQLite, inQdrant int
	var createdAt, updatedAt string
	if err := row.Scan(&it.ID, &it.Kind, &it.Title, &it.Body, &tagsRaw, &metaRaw, &inSQLite, &inQdrant, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	it.InSQLite = inSQLite != 0
	it.InQdrant = inQdrant != 0
	it.CreatedAt = parseTime(createdAt)
	it.UpdatedAt = parseTime(updatedAt)
	it.Tags = []string{}
	_ = json.Unmarshal([]byte(tagsRaw), &it.Tags)
	if it.Tags == nil {
		it.Tags = []string{}
	}
	it.Metadata = map[string]any{}
	_ = json.Unmarshal([]byte(metaRaw), &it.Metadata)
	return &it, nil
}

func (s *Service) Create(ctx context.Context, kind, title, body string, tags []string, metadata map[string]any, target StoreTarget) (*Item, error) {
	if !ValidKind(kind) {
		return nil, fmt.Errorf("invalid kind")
	}
	if !ValidTarget(string(target)) {
		return nil, fmt.Errorf("invalid store_target")
	}
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title required")
	}
	if tags == nil {
		tags = []string{}
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := marshalTags(tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	inSQLite := target == TargetSQLite || target == TargetBoth
	inQD := target == TargetQdrant || target == TargetBoth
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO knowledge (id, kind, title, body, tags, metadata, in_sqlite, in_qdrant, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, kind, title, body, tagsJSON, string(metaJSON), boolInt(inSQLite), boolInt(inQD), now, now)
	if err != nil {
		return nil, err
	}
	if inQD {
		vec, err := s.Embed.Embed(ctx, title+"\n"+body)
		if err != nil {
			return nil, err
		}
		if err := s.Qdrant.Upsert(ctx, id, vec, qdrant.PointPayload{
			ID: id, Kind: kind, Title: title, Snippet: snippet(body, 280), Tags: tags,
		}); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, id)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, kind, title, body, tags, metadata, in_sqlite, in_qdrant, created_at, updated_at
		FROM knowledge WHERE id = ?`, id)
	return scanItem(row)
}

func (s *Service) List(ctx context.Context, kind, q string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, kind, title, body, tags, metadata, in_sqlite, in_qdrant, created_at, updated_at FROM knowledge WHERE 1=1`
	args := []any{}
	if kind != "" {
		query += ` AND kind = ?`
		args = append(args, kind)
	}
	if q != "" {
		query += ` AND (lower(title) LIKE ? OR lower(body) LIKE ?)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like)
	}
	query += ` ORDER BY updated_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	if out == nil {
		out = []Item{}
	}
	return out, rows.Err()
}

func (s *Service) Delete(ctx context.Context, id string) error {
	it, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if it.InQdrant {
		_ = s.Qdrant.Delete(ctx, id)
	}
	_, err = s.DB.ExecContext(ctx, `DELETE FROM knowledge WHERE id = ?`, id)
	return err
}

type SearchMode string

const (
	ModeStructured SearchMode = "structured"
	ModeSemantic   SearchMode = "semantic"
	ModeHybrid     SearchMode = "hybrid"
)

func (s *Service) Search(ctx context.Context, query, kind string, mode SearchMode, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 20
	}
	switch mode {
	case ModeStructured, "":
		return s.List(ctx, kind, query, limit)
	case ModeSemantic:
		return s.semantic(ctx, query, kind, limit)
	case ModeHybrid:
		sem, err := s.semantic(ctx, query, kind, limit)
		if err != nil {
			return nil, err
		}
		str, err := s.List(ctx, kind, query, limit)
		if err != nil {
			return nil, err
		}
		return mergeByID(sem, str, limit), nil
	default:
		return nil, fmt.Errorf("invalid mode")
	}
}

func (s *Service) semantic(ctx context.Context, query, kind string, limit int) ([]Item, error) {
	vec, err := s.Embed.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	hits, err := s.Qdrant.Search(ctx, vec, limit, kind)
	if err != nil {
		return nil, err
	}
	out := make([]Item, 0, len(hits))
	for _, h := range hits {
		score := h.Score
		it, err := s.Get(ctx, h.ID)
		if err != nil {
			title, _ := h.Payload["title"].(string)
			snip, _ := h.Payload["snippet"].(string)
			k, _ := h.Payload["kind"].(string)
			out = append(out, Item{ID: h.ID, Kind: k, Title: title, Body: snip, Tags: []string{}, Metadata: map[string]any{}, InQdrant: true, Score: &score})
			continue
		}
		it.Score = &score
		out = append(out, *it)
	}
	return out, nil
}

func mergeByID(a, b []Item, limit int) []Item {
	seen := map[string]Item{}
	order := []string{}
	for _, it := range a {
		if _, ok := seen[it.ID]; !ok {
			order = append(order, it.ID)
		}
		seen[it.ID] = it
	}
	for _, it := range b {
		if _, ok := seen[it.ID]; ok {
			continue
		}
		seen[it.ID] = it
		order = append(order, it.ID)
	}
	out := make([]Item, 0, len(order))
	for _, id := range order {
		out = append(out, seen[id])
		if len(out) >= limit {
			break
		}
	}
	return out
}

type Stats struct {
	TotalKnowledge   int            `json:"total_knowledge"`
	ByKind           map[string]int `json:"by_kind"`
	SessionCount     int            `json:"session_count"`
	RecentKnowledge  []Item         `json:"recent_knowledge"`
}

func (s *Service) Stats(ctx context.Context) (*Stats, error) {
	st := &Stats{ByKind: map[string]int{
		"general_knowledge": 0,
		"expertise":         0,
		"resolved_issues":   0,
		"task_guide":        0,
	}}
	rows, err := s.DB.QueryContext(ctx, `SELECT kind, COUNT(*) FROM knowledge GROUP BY kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			return nil, err
		}
		st.ByKind[kind] = n
		st.TotalKnowledge += n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&st.SessionCount); err != nil {
		return nil, err
	}
	recent, err := s.List(ctx, "", "", 8)
	if err != nil {
		return nil, err
	}
	st.RecentKnowledge = recent
	return st, nil
}
