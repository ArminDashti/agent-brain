package sessions

import (
	"context"
	"database/sql"
)

func Persist(ctx context.Context, db *sql.DB, parsed *Parsed) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	s := parsed.Session
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (
			uuid, name, subtitle, status, unified_mode, workspace_id, workspace_path,
			context_usage_percent, input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
			model_config, total_lines_added, total_lines_removed, files_changed_count,
			created_at_ms, last_updated_at_ms, extracted_at, raw_json, updated_at
		) VALUES (
			$1::uuid, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13::jsonb, $14, $15, $16,
			$17, $18, $19, $20::jsonb, NOW()
		)
		ON CONFLICT (uuid) DO UPDATE SET
			name = EXCLUDED.name,
			subtitle = EXCLUDED.subtitle,
			status = EXCLUDED.status,
			unified_mode = EXCLUDED.unified_mode,
			workspace_id = EXCLUDED.workspace_id,
			workspace_path = EXCLUDED.workspace_path,
			context_usage_percent = EXCLUDED.context_usage_percent,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			cache_write_tokens = EXCLUDED.cache_write_tokens,
			model_config = EXCLUDED.model_config,
			total_lines_added = EXCLUDED.total_lines_added,
			total_lines_removed = EXCLUDED.total_lines_removed,
			files_changed_count = EXCLUDED.files_changed_count,
			created_at_ms = EXCLUDED.created_at_ms,
			last_updated_at_ms = EXCLUDED.last_updated_at_ms,
			extracted_at = EXCLUDED.extracted_at,
			raw_json = EXCLUDED.raw_json,
			updated_at = NOW()
	`, s.UUID, s.Name, s.Subtitle, s.Status, s.UnifiedMode, s.WorkspaceID, s.WorkspacePath,
		s.ContextUsagePercent, s.InputTokens, s.OutputTokens, s.CacheReadTokens, s.CacheWriteTokens,
		string(s.ModelConfigJSON), s.TotalLinesAdded, s.TotalLinesRemoved, s.FilesChangedCount,
		s.CreatedAtMs, s.LastUpdatedAtMs, s.ExtractedAt, string(s.RawJSON))
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM turns WHERE session_uuid = $1::uuid`, s.UUID); err != nil {
		return err
	}
	for _, t := range parsed.Turns {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO turns (
				session_uuid, bubble_id, ordinal, turn_type, text, tool_name, mcp_name, status,
				duration_ms, has_thinking, thinking_duration_ms, thinking_text,
				input_tokens, output_tokens, cache_read_tokens, cache_write_tokens, bubble_json
			) VALUES (
				$1::uuid, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, $11, $12,
				$13, $14, $15, $16, $17::jsonb
			)
		`, s.UUID, t.BubbleID, t.Ordinal, t.TurnType, t.Text, t.ToolName, t.MCPName, t.Status,
			t.DurationMs, t.HasThinking, t.ThinkingDurationMs, t.ThinkingText,
			t.InputTokens, t.OutputTokens, t.CacheReadTokens, t.CacheWriteTokens, string(t.BubbleJSON)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
