#Requires -Version 5.1
<#
.SYNOPSIS
  Remove agent-brain-mcp completely from this Windows machine.

.DESCRIPTION
  Unregisters the Cursor MCP entry, deletes node_modules and dist.
  Leaves source and .env in the repo.

.EXAMPLE
  .\remove-win.ps1
#>
[CmdletBinding()]
param(
  [string]$McpRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path,
  [string]$ServerName = 'agent-brain',
  [string]$CursorMcpPath = (Join-Path $env:USERPROFILE '.cursor\mcp.json'),
  [switch]$RemoveEnv
)

$ErrorActionPreference = 'Stop'

function Write-Step([string]$Message) {
  Write-Host "==> $Message"
}

function Remove-CursorMcpEntry {
  if (-not (Test-Path -LiteralPath $CursorMcpPath)) {
    Write-Step "No Cursor MCP config at $CursorMcpPath"
    return
  }

  $backup = "$CursorMcpPath.bak-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
  Copy-Item -LiteralPath $CursorMcpPath -Destination $backup -Force
  Write-Step "Backed up Cursor MCP config → $backup"

  # Node parses case-sensitive JSON keys; ConvertFrom-Json fails on NO_PROXY/no_proxy.
  $result = & node -e "const fs=require('fs');const p=process.argv[1];const n=process.argv[2];if(!fs.existsSync(p)){process.stdout.write('missing');process.exit(0)}const cfg=JSON.parse(fs.readFileSync(p,'utf8'));if(!cfg.mcpServers){process.stdout.write('no-servers');process.exit(0)}if(!Object.prototype.hasOwnProperty.call(cfg.mcpServers,n)){process.stdout.write('absent');process.exit(0)}delete cfg.mcpServers[n];fs.writeFileSync(p,JSON.stringify(cfg,null,2)+'\n','utf8');process.stdout.write('removed')" $CursorMcpPath $ServerName
  if ($LASTEXITCODE -ne 0) { throw "Failed to update Cursor MCP config ($LASTEXITCODE)" }

  switch ($result) {
    'no-servers' { Write-Step "No mcpServers object in $CursorMcpPath" }
    'absent'     { Write-Step "MCP '$ServerName' not registered" }
    'removed'    { Write-Step "Removed Cursor MCP '$ServerName'" }
    default      { Write-Step "No Cursor MCP config at $CursorMcpPath" }
  }
}

function Remove-Tree([string]$Path) {
  if (-not (Test-Path -LiteralPath $Path)) {
    Write-Step "Skip (missing): $Path"
    return
  }
  Write-Step "Removing $Path"
  Remove-Item -LiteralPath $Path -Recurse -Force
}

Write-Step 'Removing agent-brain-mcp'
Remove-CursorMcpEntry
Remove-Tree (Join-Path $McpRoot 'node_modules')
Remove-Tree (Join-Path $McpRoot 'dist')

if ($RemoveEnv) {
  Remove-Tree (Join-Path $McpRoot '.env')
}

Write-Host "Done: agent-brain-mcp removed (reload Cursor if it was running)"
