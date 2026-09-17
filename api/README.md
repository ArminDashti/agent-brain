# agent-brain API

Gin + SQLite + Qdrant knowledge API.

```bash
cp .env.example .env
docker compose up -d   # Qdrant only
go mod tidy
go run ./cmd/server
```

Env: `SQLITE_PATH` (default `./data/agent_brain.db`), `QDRANT_URL`, `JWT_SECRET`.

Login: `armin` / `dopadopa123` — API `:8220`.

Kinds: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.  
`store_target`: `sqlite` | `qdrant` | `both`.
