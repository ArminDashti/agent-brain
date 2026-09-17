---
name: ingest-agent-session
description: >-
  Extracts Cursor composer session(s) from state.vscdb and ingests
  context %, tokens, thinking, and turn metadata into agent-brain Sessions
  (SQLite sessions/turns).
disable-model-invocation: false
metadata:
  version: "1.2.0"
  author: Armin Dashti
  category: logging
  tags: [cursor, sessions, tokens, thinking, ingest, agent-brain]
  last_updated: "2026-09-17 15:50:00"
  uuid: a1b2c3d4-e5f6-4789-a012-3456789abcde
---

# Ingest Agent Session

## Objective

Persist Cursor session extract(s) into agent-brain SQLite
(`sessions` / `turns`) via agent-brain-api, prioritizing **context**,
**token counts**, **thinking**, and useful turn metadata. WebUI **Sessions**
browses this data.

## When

- After every agent response (required by alwaysApply rule)
- When the user asks to sync/ingest the current session
- When the user asks to refill all Sessions from Cursor (`sync`)

## How

1. **Resolve UUID** (single ingest)
   - Prefer the active composer/chat UUID from the conversation or
     agent-transcripts folder name.
   - If unknown, stop with a one-line fail: `ltm-ingest: missing session uuid`.

2. **Prefer dispatcher (extract + API ingest in one step)**
   - Resolve binary:
     - `$env:AGENT_BRAIN_DISPATCHER` if set
     - else `C:\Users\armin\GitHub\agent-brain\dispatcher-win\agent-brain-dispatcher.exe`
   - Single session (PowerShell):

```powershell
$bin = if ($env:AGENT_BRAIN_DISPATCHER) { $env:AGENT_BRAIN_DISPATCHER } else { "C:\Users\armin\GitHub\agent-brain\dispatcher-win\agent-brain-dispatcher.exe" }
& $bin ingest --uuid <UUID>
```

   - All composers from `state.vscdb` (fills WebUI Sessions):

```powershell
$bin = if ($env:AGENT_BRAIN_DISPATCHER) { $env:AGENT_BRAIN_DISPATCHER } else { "C:\Users\armin\GitHub\agent-brain\dispatcher-win\agent-brain-dispatcher.exe" }
& $bin sync
```

   - Default DB: `%APPDATA%\Cursor\User\globalStorage\state.vscdb`
   - Success: one JSON line (`ingest`: `ok`/`uuid`/tokens; `sync`: `ok`/`synced`/`failed`/`total`).
   - Report: `ltm-ingest: ok <uuid> context=<pct> in=<n> out=<n>` or `ltm-ingest: sync synced=<n> failed=<n>` and stop.

3. **Fallback if binary missing**
   - Extract via PowerShell script:

```powershell
& "C:\Users\armin\GitHub\armin-command-center\cursor-db\extract-session-logs.ps1" -Uuid <UUID> -Format json
```

   - Output: `...\cursor-db\sessions\<uuid>.json`
   - Prefer MCP `ingest_session(file_path="...")` when available; else HTTP login +
     `POST /api/v1/sessions/ingest`.

4. **Report**
   - Success: `ltm-ingest: ok <uuid> context=<pct> in=<n> out=<n>`
   - Failure: `ltm-ingest: fail <reason>` (one line). Continue the turn; do not invent data.

## Always

1. **Always** ingest after the agent response when this skill/rule is loaded.
2. **Always** prefer `agent-brain-dispatcher.exe` when present.
3. **Always** keep secrets out of chat (use env defaults).

## Never

1. **Never** skip ingest because the turn “looks small”.
2. **Never** commit `.env` or paste passwords into the chat.
3. **Never** rewrite `state.vscdb` — read-only extract only.
4. **Never** use centralized-agent-data MCP/API for new ingests — use agent-brain.
