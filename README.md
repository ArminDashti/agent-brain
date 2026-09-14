# agent-brain

Centralized store for agent experience: general knowledge, solutions, task procedures, and expertise.

**Separate project** — own remote, own local tree, own Postgres + Qdrant. Not nested under other repos.

## Layout

| Folder | Role |
|--------|------|
| `agent-brain-api` | Gin API + PostgreSQL + Qdrant |
| `agent-brain-webui` | Vue + Tailwind + shadcn + Inter + PWA |
| `agent-brain-mcp` | MCP stdio server (agents talk to the API) |
| `agent-brain-rule-skills` | Cursor plugin: rules + skills for store/search |

## Quick start

```bash
cd agent-brain-api && cp .env.example .env && docker compose up -d && go run ./cmd/server
cd ../agent-brain-webui && cp .env.example .env && npm install && npm run dev
cd ../agent-brain-mcp && cp .env.example .env && npm install && npm run build
```

Default login: `armin` / `dopadopa123`

Ports: Postgres `5470`, Qdrant `6335`/`6336`, API `8220`, WebUI `5220`.

## Storage choice

Agents pick `store_target` per write: `postgres` (structured/filterable), `qdrant` (semantic), or `both`.
