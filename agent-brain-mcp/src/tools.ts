import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { z } from 'zod'
import { BrainClient } from './client.js'

function textResult(data: unknown) {
  return { content: [{ type: 'text' as const, text: JSON.stringify(data, null, 2) }] }
}

function errorResult(err: unknown) {
  const message = err instanceof Error ? err.message : String(err)
  return { content: [{ type: 'text' as const, text: message }], isError: true as const }
}

const kindSchema = z.enum(['general_knowledge', 'solution', 'task_procedure', 'expertise'])
const targetSchema = z.enum(['postgres', 'qdrant', 'both'])
const modeSchema = z.enum(['structured', 'semantic', 'hybrid'])

export function registerTools(server: McpServer, client: BrainClient): void {
  server.tool('health_check', 'Check agent-brain-api health (Postgres + Qdrant).', {}, async () => {
    try {
      return textResult(await client.health())
    } catch (err) {
      return errorResult(err)
    }
  })

  server.tool(
    'login',
    'Login and cache JWT. Optional username/password override env defaults.',
    {
      username: z.string().optional(),
      password: z.string().optional(),
    },
    async ({ username, password }) => {
      try {
        const result = await client.login(username, password)
        return textResult({
          ok: true,
          user: result.user,
          token_preview: `${String((result as { token: string }).token).slice(0, 12)}…`,
        })
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'store_knowledge',
    'Store knowledge. Prefer postgres for exact/filterable records, qdrant for semantic recall, both when listable and searchable.',
    {
      kind: kindSchema,
      title: z.string(),
      body: z.string().default(''),
      tags: z.array(z.string()).optional(),
      store_target: targetSchema.describe('postgres | qdrant | both'),
      metadata: z.record(z.unknown()).optional(),
    },
    async (args) => {
      try {
        return textResult(await client.storeKnowledge(args))
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'search_knowledge',
    'Search knowledge. structured=Postgres ILIKE, semantic=Qdrant vectors, hybrid=merge both.',
    {
      query: z.string(),
      kind: kindSchema.optional(),
      mode: modeSchema.default('hybrid'),
      limit: z.number().int().positive().optional(),
    },
    async (args) => {
      try {
        return textResult(await client.searchKnowledge(args))
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'list_knowledge',
    'List knowledge rows from Postgres (optional kind/q filter).',
    {
      kind: kindSchema.optional(),
      q: z.string().optional(),
    },
    async ({ kind, q }) => {
      try {
        return textResult(await client.listKnowledge(kind, q))
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'get_knowledge',
    'Get one knowledge item by id.',
    { id: z.string().uuid() },
    async ({ id }) => {
      try {
        return textResult(await client.getKnowledge(id))
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'delete_knowledge',
    'Delete knowledge from Postgres and Qdrant (if present).',
    { id: z.string().uuid() },
    async ({ id }) => {
      try {
        return textResult(await client.deleteKnowledge(id))
      } catch (err) {
        return errorResult(err)
      }
    },
  )
}
