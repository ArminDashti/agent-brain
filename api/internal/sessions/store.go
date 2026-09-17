package sessions

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func Persist(ctx context.Context, db *sql.DB, parsed *Parsed) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	s := parsed.Session
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var extracted any
	if s.ExtractedAt != nil {
		extracted = s.ExtractedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (
			uuid, name, subtitle, status, unified_mode, workspace_id, workspace_path,
			context_usage_percent, input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
			model_config, total_lines_added, total_lines_removed, files_changed_count,
			created_at_ms, last_updated_at_ms, extracted_at, raw_json, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
		ON CONFLICT (uuid) DO UPDATE SET
			name = excluded.name,
			subtitle = excluded.subtitle,
			status = excluded.status,
			unified_mode = excluded.unified_mode,
			workspace_id = excluded.workspace_id,
			workspace_path = excluded.workspace_path,
			context_usage_percent = excluded.context_usage_percent,
			input_tokens = excluded.input_tokens,
			output_tokens = excluded.output_tokens,
			cache_read_tokens = excluded.cache_read_tokens,
			cache_write_tokens = excluded.cache_write_tokens,
			model_config = excluded.model_config,
			total_lines_added = excluded.total_lines_added,
			total_lines_removed = excluded.total_lines_removed,
			files_changed_count = excluded.files_changed_count,
			created_at_ms = excluded.created_at_ms,
			last_updated_at_ms = excluded.last_updated_at_ms,
			extracted_at = excluded.extracted_at,
			raw_json = excluded.raw_json,
			updated_at = excluded.updated_at
	`, s.UUID, s.Name, s.Subtitle, s.Status, s.UnifiedMode, s.WorkspaceID, s.WorkspacePath,
		s.ContextUsagePercent, s.InputTokens, s.OutputTokens, s.CacheReadTokens, s.CacheWriteTokens,
		string(s.ModelConfigJSON), s.TotalLinesAdded, s.TotalLinesRemoved, s.FilesChangedCount,
		s.CreatedAtMs, s.LastUpdatedAtMs, extracted, string(s.RawJSON), now)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM turns WHERE session_uuid = ?`, s.UUID); err != nil {
		return err
	}
	for _, t := range parsed.Turns {
		id := uuid.NewString()
		hasThinking := 0
		if t.HasThinking {
			hasThinking = 1
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO turns (
				id, session_uuid, bubble_id, ordinal, turn_type, text, tool_name, mcp_name, status,
				duration_ms, has_thinking, thinking_duration_ms, thinking_text,
				input_tokens, output_tokens, cache_read_tokens, cache_write_tokens, bubble_json
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?,
				?, ?, ?, ?, ?
			)
		`, id, s.UUID, t.BubbleID, t.Ordinal, t.TurnType, t.Text, t.ToolName, t.MCPName, t.Status,
			t.DurationMs, hasThinking, t.ThinkingDurationMs, t.ThinkingText,
			t.InputTokens, t.OutputTokens, t.CacheReadTokens, t.CacheWriteTokens, string(t.BubbleJSON)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
