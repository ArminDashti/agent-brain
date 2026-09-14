---
name: search-agent-brain
description: Search agent-brain for prior solutions, procedures, and expertise via MCP.
---

# Search agent-brain

Use MCP tool `search_knowledge` on `agent-brain-mcp`.

## Modes

| Mode | Backend |
|------|---------|
| `structured` | PostgreSQL ILIKE / filters |
| `semantic` | Qdrant vector search |
| `hybrid` | Merge both (default) |

Optional `kind` filter: `general_knowledge`, `solution`, `task_procedure`, `expertise`.

Also available: `list_knowledge`, `get_knowledge`.
