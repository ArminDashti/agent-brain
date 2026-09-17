# agent-brain-mcp

Stdio MCP server. Agents call tools that hit **agent-brain API** over HTTP.

## Setup

```bash
cp .env.example .env
npm install
npm run build
```

Cursor MCP example (API published on Docker host port 5090):

```json
{
  "agent-brain": {
    "command": "node",
    "args": ["C:/Users/armin/GitHub/agent-brain/mcp/dist/index.js"],
    "env": {
      "AGENT_BRAIN_API_URL": "http://127.0.0.1:5090",
      "AGENT_BRAIN_USERNAME": "armin",
      "AGENT_BRAIN_PASSWORD": "dopadopa123"
    }
  }
}
```

## Tools

- `health_check`, `login`
- `store_knowledge` (kind, title, body, tags, store_target: sqlite|qdrant|both)
- `search_knowledge` (query, kind?, mode)
- `list_knowledge`, `get_knowledge`, `delete_knowledge`
- Session tools: `ingest_session`, `list_sessions`, `get_session_*`

Kinds: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.
