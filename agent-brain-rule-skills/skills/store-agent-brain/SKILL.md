---
name: store-agent-brain
description: Store experience, solutions, procedures, or expertise into agent-brain via MCP.
---

# Store into agent-brain

Use MCP tool `store_knowledge` on `agent-brain-mcp`.

## Fields

| Field | Values |
|-------|--------|
| `kind` | `general_knowledge` \| `solution` \| `task_procedure` \| `expertise` |
| `title` | Short label |
| `body` | Full text |
| `tags` | Optional string array |
| `store_target` | `postgres` \| `qdrant` \| `both` |

## Target choice

- Structured inventory / filters → `postgres`
- “Find similar later” → `qdrant` or `both`
- Default for important fixes → `both`

Login uses env defaults (`armin` / `dopadopa123`) unless you call `login` first.
