CREATE TABLE IF NOT EXISTS sessions (
    uuid TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    subtitle TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    unified_mode TEXT NOT NULL DEFAULT '',
    workspace_id TEXT NOT NULL DEFAULT '',
    workspace_path TEXT NOT NULL DEFAULT '',
    context_usage_percent REAL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    cache_read_tokens INTEGER,
    cache_write_tokens INTEGER,
    model_config TEXT NOT NULL DEFAULT '{}',
    total_lines_added INTEGER NOT NULL DEFAULT 0,
    total_lines_removed INTEGER NOT NULL DEFAULT 0,
    files_changed_count INTEGER NOT NULL DEFAULT 0,
    created_at_ms INTEGER,
    last_updated_at_ms INTEGER,
    extracted_at TEXT,
    raw_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_workspace ON sessions (workspace_id);
CREATE INDEX IF NOT EXISTS idx_sessions_name ON sessions (name);

CREATE TABLE IF NOT EXISTS turns (
    id TEXT PRIMARY KEY,
    session_uuid TEXT NOT NULL REFERENCES sessions (uuid) ON DELETE CASCADE,
    bubble_id TEXT NOT NULL,
    ordinal INTEGER NOT NULL,
    turn_type TEXT NOT NULL DEFAULT '',
    text TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL DEFAULT '',
    mcp_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER,
    has_thinking INTEGER NOT NULL DEFAULT 0,
    thinking_duration_ms INTEGER,
    thinking_text TEXT NOT NULL DEFAULT '',
    input_tokens INTEGER,
    output_tokens INTEGER,
    cache_read_tokens INTEGER,
    cache_write_tokens INTEGER,
    bubble_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (session_uuid, bubble_id)
);

CREATE INDEX IF NOT EXISTS idx_turns_session_ordinal ON turns (session_uuid, ordinal);
CREATE INDEX IF NOT EXISTS idx_turns_thinking ON turns (session_uuid, has_thinking);
