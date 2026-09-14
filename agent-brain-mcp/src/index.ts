#!/usr/bin/env node
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { BrainClient } from './client.js'
import { registerTools } from './tools.js'

async function main() {
  const client = new BrainClient()
  const server = new McpServer({
    name: 'agent-brain-mcp',
    version: '1.0.0',
  })
  registerTools(server, client)
  const transport = new StdioServerTransport()
  await server.connect(transport)
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
