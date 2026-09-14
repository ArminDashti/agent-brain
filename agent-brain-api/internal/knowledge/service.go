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
	"github.com/jackc/pgx/v5/pgtype"
)

type StoreTarget string

const (
	TargetPostgres StoreTarget = "postgres"
	TargetQdrant   StoreTarget = "qdrant"
	TargetBoth     StoreTarget = "both"
)

func ValidKind(k string) bool {
	switch k {
	case "general_knowledge", "solution", "task_procedure", "expertise":
		return true
	}
	return false
}

func ValidTarget(t string) bool {
	switch StoreTarget(t) {
	case TargetPostgres, TargetQdrant, TargetBoth:
		return true
	}
	return false
}

type Item struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	Tags       []string       `json:"tags"`
	Metadata   map[string]any `json:"metadata"`
	InPostgres bool           `json:"in_postgres"`
	InQdrant   bool           `json:"in_qdrant"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Score      *float64       `json:"score,omitempty"`
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

func tagsValue(tags []string) any {
	if tags == nil {
		tags = []string{}
	}
	return pgtype.FlatArray[string](tags)
}

type scannable interface{ Scan(dest ...any) error }

func scanItem(row scannable) (*Item, error) {
	var it Item
	var tags pgtype.FlatArray[string]
	var metaRaw string
	if err := row.Scan(&it.ID, &it.Kind, &it.Title, &it.Body, &tags, &metaRaw, &it.InPostgres, &it.InQdrant, &it.CreatedAt, &it.UpdatedAt); err != nil {
		return nil, err
	}
	it.Tags = []string(tags)
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
	id := uuid.NewString()
	inPG := target == TargetPostgres || target == TargetBoth
	inQD := target == TargetQdrant || target == TargetBoth
	now := time.Now().UTC()
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO knowledge (id, kind, title, body, tags, metadata, in_postgres, in_qdrant, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $9)`,
		id, kind, title, body, tagsValue(tags), string(metaJSON), inPG, inQD, now)
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

func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id::text, kind, title, body, tags, metadata::text, in_postgres, in_qdrant, created_at, updated_at
		FROM knowledge WHERE id = $1::uuid`, id)
	return scanItem(row)
}

func (s *Service) List(ctx context.Context, kind, q string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id::text, kind, title, body, tags, metadata::text, in_postgres, in_qdrant, created_at, updated_at FROM knowledge WHERE 1=1`
	args := []any{}
	n := 1
	if kind != "" {
		query += fmt.Sprintf(" AND kind = $%d", n)
		args = append(args, kind)
		n++
	}
	if q != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR body ILIKE $%d)", n, n)
		args = append(args, "%"+q+"%")
		n++
	}
	query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d", n)
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
	_, err = s.DB.ExecContext(ctx, `DELETE FROM knowledge WHERE id = $1::uuid`, id)
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
