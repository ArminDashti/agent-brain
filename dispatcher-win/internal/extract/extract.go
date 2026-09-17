package extract

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

// Session pulls Cursor composer/chat data from state.vscdb into the ingest JSON shape.
func Session(uuid, dbPath string) ([]byte, error) {
	uuid = strings.ToLower(strings.TrimSpace(uuid))
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
	}
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("DB not found: %s", abs)
	}

	db, tmpCopy, err := openDB(abs)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
		if tmpCopy != "" {
			for _, p := range []string{tmpCopy, tmpCopy + "-wal", tmpCopy + "-shm"} {
				_ = os.Remove(p)
			}
		}
	}()

	header, err := loadHeader(db, uuid)
	if err != nil {
		return nil, err
	}

	diskRows, err := queryKV(db, "SELECT key, value FROM cursorDiskKV WHERE key LIKE ?", "%"+uuid+"%")
	if err != nil {
		return nil, err
	}
	itemRows, err := queryKV(db, "SELECT key, value FROM ItemTable WHERE key LIKE ?", "%"+uuid+"%")
	if err != nil {
		return nil, err
	}

	var composerData any
	bubblesByID := map[string]any{}
	checkpoints := []map[string]any{}
	otherDisk := []map[string]any{}
	ofsContent := []map[string]any{}

	prefixComposer := "composerdata:" + uuid
	prefixBubble := "bubbleid:" + uuid + ":"
	prefixCheckpoint := "checkpointid:" + uuid + ":"

	for _, row := range diskRows {
		key := row.key
		val := row.value
		lk := strings.ToLower(key)
		switch {
		case lk == prefixComposer:
			composerData = val
		case strings.HasPrefix(lk, prefixBubble):
			parts := strings.SplitN(key, ":", 3)
			bid := ""
			if len(parts) == 3 {
				bid = parts[2]
			}
			if m, ok := val.(map[string]any); ok {
				cp := copyMap(m)
				cp["_key"] = key
				bubblesByID[bid] = cp
			} else {
				bubblesByID[bid] = val
			}
		case strings.HasPrefix(lk, prefixCheckpoint):
			checkpoints = append(checkpoints, map[string]any{"key": key, "value": val})
		case strings.HasPrefix(lk, "ofscontent:") || strings.Contains(lk, "ofscontent"):
			ofsContent = append(ofsContent, map[string]any{"key": key, "value": val})
		default:
			otherDisk = append(otherDisk, map[string]any{"key": key, "value": val})
		}
	}

	itemEntries := make([]map[string]any, 0, len(itemRows))
	for _, r := range itemRows {
		itemEntries = append(itemEntries, map[string]any{"key": r.key, "value": r.value})
	}

	ordered := []map[string]any{}
	if cd, ok := composerData.(map[string]any); ok {
		if headersOnly, ok := cd["fullConversationHeadersOnly"].([]any); ok {
			for _, h := range headersOnly {
				hm, ok := h.(map[string]any)
				if !ok {
					continue
				}
				bid, _ := hm["bubbleId"].(string)
				if bid == "" {
					continue
				}
				bubble := bubblesByID[bid]
				delete(bubblesByID, bid)
				ordered = append(ordered, map[string]any{
					"bubbleId": bid,
					"header":   hm,
					"bubble":   bubble,
				})
			}
		}
	}
	for bid, bubble := range bubblesByID {
		ordered = append(ordered, map[string]any{
			"bubbleId": bid,
			"header":   nil,
			"bubble":   bubble,
		})
	}

	found := header != nil || composerData != nil || len(ordered) > 0 || len(itemEntries) > 0 || len(otherDisk) > 0
	if !found {
		return nil, fmt.Errorf("no rows found for UUID %s", uuid)
	}

	summary := buildSummary(composerData, header, len(ordered), len(checkpoints), len(itemEntries))

	result := map[string]any{
		"uuid":          uuid,
		"extractedAt":   time.Now().UTC().Format(time.RFC3339Nano),
		"dbPath":        abs,
		"found":         true,
		"summary":       summary,
		"header":        header,
		"composerData":  composerData,
		"conversation":  ordered,
		"checkpoints":   checkpoints,
		"ofsContent":    ofsContent,
		"itemTable":     itemEntries,
		"otherDiskKeys": otherDisk,
	}

	return json.Marshal(result)
}

// ComposerRef is a session/composer id discovered in state.vscdb.
type ComposerRef struct {
	UUID string `json:"uuid"`
	Name string `json:"name,omitempty"`
}

// ListComposers returns all composer UUIDs from state.vscdb (composerHeaders, else cursorDiskKV).
func ListComposers(dbPath string) ([]ComposerRef, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
	}
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("DB not found: %s", abs)
	}

	db, tmpCopy, err := openDB(abs)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
		if tmpCopy != "" {
			for _, p := range []string{tmpCopy, tmpCopy + "-wal", tmpCopy + "-shm"} {
				_ = os.Remove(p)
			}
		}
	}()

	seen := map[string]ComposerRef{}

	rows, err := db.Query(`SELECT composerId FROM composerHeaders`)
	if err == nil {
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			seen[id] = ComposerRef{UUID: id}
		}
		scanErr := rows.Err()
		_ = rows.Close()
		if scanErr != nil {
			return nil, scanErr
		}
	} else if !strings.Contains(strings.ToLower(err.Error()), "no such table") {
		return nil, err
	}

	if len(seen) == 0 {
		kvRows, err := queryKV(db, `SELECT key, value FROM cursorDiskKV WHERE lower(key) LIKE 'composerdata:%'`)
		if err != nil {
			return nil, err
		}
		for _, row := range kvRows {
			parts := strings.SplitN(row.key, ":", 2)
			if len(parts) != 2 {
				continue
			}
			id := strings.ToLower(strings.TrimSpace(parts[1]))
			if id == "" {
				continue
			}
			ref := ComposerRef{UUID: id}
			if m, ok := row.value.(map[string]any); ok {
				if n, ok := m["name"].(string); ok {
					ref.Name = n
				}
			}
			seen[id] = ref
		}
	}

	out := make([]ComposerRef, 0, len(seen))
	for _, ref := range seen {
		out = append(out, ref)
	}
	return out, nil
}

func buildSummary(composerData any, header map[string]any, bubbleCount, checkpointCount, itemCount int) map[string]any {
	cd, _ := composerData.(map[string]any)
	getCD := func(k string) any {
		if cd == nil {
			return nil
		}
		return cd[k]
	}
	getH := func(k string) any {
		if header == nil {
			return nil
		}
		return header[k]
	}

	createdAt := getCD("createdAt")
	if createdAt == nil {
		createdAt = getH("createdAt")
	}
	lastUpdatedAt := getCD("lastUpdatedAt")
	if lastUpdatedAt == nil {
		lastUpdatedAt = getH("lastUpdatedAt")
	}

	return map[string]any{
		"name":             getCD("name"),
		"subtitle":         getCD("subtitle"),
		"status":           getCD("status"),
		"unifiedMode":      getCD("unifiedMode"),
		"createdAt":        createdAt,
		"lastUpdatedAt":    lastUpdatedAt,
		"workspaceId":      getH("workspaceId"),
		"isArchived":       getH("isArchived"),
		"isSubagent":       getH("isSubagent"),
		"subagentTypeName": getH("subagentTypeName"),
		"bubbleCount":      bubbleCount,
		"checkpointCount":  checkpointCount,
		"itemTableCount":   itemCount,
		"modelConfig":      getCD("modelConfig"),
	}
}

func openDB(path string) (*sql.DB, string, error) {
	uri := "file:" + filepath.ToSlash(path) + "?mode=ro"
	db, err := sql.Open("sqlite", uri)
	if err == nil {
		if pingErr := db.Ping(); pingErr == nil {
			return db, "", nil
		}
		_ = db.Close()
	}

	tmp, err := os.CreateTemp("", "state-vscdb-*.vscdb")
	if err != nil {
		return nil, "", err
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	if err := copyFile(path, tmpName); err != nil {
		_ = os.Remove(tmpName)
		return nil, "", err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		side := path + suffix
		if _, err := os.Stat(side); err == nil {
			_ = copyFile(side, tmpName+suffix)
		}
	}
	db, err = sql.Open("sqlite", tmpName)
	if err != nil {
		cleanupTemp(tmpName)
		return nil, "", err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		cleanupTemp(tmpName)
		return nil, "", err
	}
	return db, tmpName, nil
}

func cleanupTemp(tmpName string) {
	for _, p := range []string{tmpName, tmpName + "-wal", tmpName + "-shm"} {
		_ = os.Remove(p)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func loadHeader(db *sql.DB, uuid string) (map[string]any, error) {
	rows, err := db.Query("SELECT * FROM composerHeaders WHERE lower(composerId)=?", uuid)
	if err != nil {
		// Table may be missing on older Cursor DBs; treat as no header.
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}
	header := map[string]any{}
	for i, col := range cols {
		header[col] = decodeSQLValue(vals[i])
	}
	if v, ok := header["value"]; ok {
		header["value"] = decodeValue(v)
	}
	return header, nil
}

type kvRow struct {
	key   string
	value any
}

func queryKV(db *sql.DB, query string, args ...any) ([]kvRow, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := []kvRow{}
	for rows.Next() {
		var key string
		var raw any
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		out = append(out, kvRow{key: key, value: decodeValue(raw)})
	}
	return out, rows.Err()
}

func decodeSQLValue(raw any) any {
	switch v := raw.(type) {
	case nil:
		return nil
	case []byte:
		return string(v)
	default:
		return v
	}
}

func decodeValue(raw any) any {
	if raw == nil {
		return nil
	}
	var text string
	switch v := raw.(type) {
	case []byte:
		if !utf8.Valid(v) {
			return map[string]any{
				"_encoding": "base64",
				"data":      base64.StdEncoding.EncodeToString(v),
			}
		}
		text = string(v)
	case string:
		text = v
	default:
		text = fmt.Sprint(v)
	}
	text = strings.Trim(text, "\x00")
	if text == "" {
		return ""
	}
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err == nil {
		return parsed
	}
	return text
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m)+1)
	for k, v := range m {
		out[k] = v
	}
	return out
}
