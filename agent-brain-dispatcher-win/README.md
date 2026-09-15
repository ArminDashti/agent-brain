# agent-brain-dispatcher-win

Windows Go CLI that extracts the current Cursor composer session from
`state.vscdb` and ingests it into agent-brain via HTTP API.

## Build

```powershell
cd C:\Users\armin\GitHub\agent-brain\agent-brain-dispatcher-win
go build -o agent-brain-dispatcher.exe ./cmd/agent-brain-dispatcher
```

## Usage

```powershell
.\agent-brain-dispatcher.exe ingest --uuid <session-uuid>
```

Optional flags:

| Flag | Default |
|------|---------|
| `--db` | `%APPDATA%\Cursor\User\globalStorage\state.vscdb` |
| `--api-url` | `http://127.0.0.1:8220` |
| `--out` | (none) write extract JSON for debug |
| `--username` / `--password` | env or `armin` / `dopadopa123` |

## Environment

| Variable | Purpose |
|----------|---------|
| `AGENT_BRAIN_API_URL` | API base URL |
| `AGENT_BRAIN_USERNAME` | Login user |
| `AGENT_BRAIN_PASSWORD` | Login password |
| `AGENT_BRAIN_STATE_VSCDB` | Override path to `state.vscdb` |
| `AGENT_BRAIN_DISPATCHER` | Absolute path to this exe (for skills) |

## Output

Success (stdout, one line):

```json
{"ok":true,"uuid":"...","turn_count":12,"context_pct":41.2,"input_tokens":1000,"output_tokens":500}
```

Failure (non-zero exit):

```json
{"ok":false,"error":"..."}
```
