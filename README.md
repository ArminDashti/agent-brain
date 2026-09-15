# agent-brain

Centralized store for agent experience: general knowledge, solutions, task procedures, expertise, and **Long term memory** (Cursor agent sessions).

**Separate project** — own remote, own local tree, own Postgres + Qdrant. Not nested under other repos.

## Layout

| Folder | Role |
|--------|------|
| `agent-brain-api` | Gin API + PostgreSQL + Qdrant |
| `agent-brain-webui` | Vue + Tailwind + shadcn + Inter + PWA |
| `agent-brain-mcp` | MCP stdio server (agents talk to the API) |
| `agent-brain-rule-skills` | Cursor plugin: rules + skills for store/search/ingest |
| `agent-brain-dispatcher-win` | Windows CLI: extract Cursor session → API ingest |

## Two surfaces

| Surface | Storage | Purpose |
|---------|---------|---------|
| **Knowledge** | Postgres `knowledge` + optional Qdrant | Durable solutions, procedures, expertise |
| **Long term memory** | Postgres `sessions` + `turns` only | Cursor session context %, tokens, thinking, turns |

WebUI: knowledge kinds + Search, plus **Long term memory** (`/long-term-memory`).

MCP: `store_knowledge` / `search_knowledge` / … and `ingest_session` / `list_sessions` / `get_session_*`.

## Quick start

```bash
cd agent-brain-api && cp .env.example .env && docker compose up -d && go run ./cmd/server
cd ../agent-brain-webui && cp .env.example .env && npm install && npm run dev
cd ../agent-brain-mcp && cp .env.example .env && npm install && npm run build
```

Default login: `armin` / `dopadopa123`

Ports: Postgres `5470`, Qdrant `6335`/`6336`, API `8220`, WebUI `5220`.

## Storage choice (knowledge)

Agents pick `store_target` per write: `postgres` (structured/filterable), `qdrant` (semantic), or `both`.

Session ingest always writes to PostgreSQL (no Qdrant).
