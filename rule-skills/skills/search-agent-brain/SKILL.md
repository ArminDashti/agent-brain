---
name: search-agent-brain
description: Search agent-brain for prior resolved issues, task guides, and expertise via MCP.
---

# Search agent-brain

Use MCP tool `search_knowledge` on `agent-brain-mcp`.

## Modes

| Mode | Backend |
|------|---------|
| `structured` | SQLite LIKE / filters |
| `semantic` | Qdrant vector search |
| `hybrid` | Merge both (default) |

Optional `kind` filter: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.

Also available: `list_knowledge`, `get_knowledge`.
