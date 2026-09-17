---
name: store-agent-brain
description: Store experience, resolved issues, task guides, or expertise into agent-brain via MCP.
---

# Store into agent-brain

Use MCP tool `store_knowledge` on `agent-brain-mcp`.

## Fields

| Field | Values |
|-------|--------|
| `kind` | `general_knowledge` \| `expertise` \| `resolved_issues` \| `task_guide` |
| `title` | Short label |
| `body` | Full text |
| `tags` | Optional string array |
| `store_target` | `sqlite` \| `qdrant` \| `both` |

## Target choice

- Structured inventory / filters → `sqlite`
- “Find similar later” → `qdrant` or `both`
- Default for important fixes → `both`

Login uses env defaults (`armin` / `dopadopa123`) unless you call `login` first.
