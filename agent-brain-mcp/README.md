# agent-brain-mcp

Stdio MCP server. Agents call tools that hit **agent-brain-api** over HTTP.

## Setup

```bash
cp .env.example .env
npm install
npm run build
```

Cursor MCP example:

```json
{
  "agent-brain": {
    "command": "node",
    "args": ["C:/Users/armin/GitHub/agent-brain/agent-brain-mcp/dist/index.js"],
    "env": {
      "AGENT_BRAIN_API_URL": "http://127.0.0.1:8220",
      "AGENT_BRAIN_USERNAME": "armin",
      "AGENT_BRAIN_PASSWORD": "dopadopa123"
    }
  }
}
```

## Tools

- `health_check`, `login`
- `store_knowledge` (kind, title, body, tags, store_target)
- `search_knowledge` (query, kind?, mode)
- `list_knowledge`, `get_knowledge`, `delete_knowledge`
