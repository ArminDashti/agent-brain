#Requires -Version 5.1
<#
.SYNOPSIS
  Install or update agent-brain-mcp on this Windows machine.

.DESCRIPTION
  Fresh install: npm install, build, ensure .env, register in Cursor mcp.json.
  Already installed: refresh deps/build and upsert the Cursor MCP entry (update).

.EXAMPLE
  .\install-win.ps1
#>
[CmdletBinding()]
param(
  [string]$McpRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path,
  [string]$ServerName = 'agent-brain',
  [string]$CursorMcpPath = (Join-Path $env:USERPROFILE '.cursor\mcp.json')
)

$ErrorActionPreference = 'Stop'

function Write-Step([string]$Message) {
  Write-Host "==> $Message"
}

function Assert-Command([string]$Name) {
  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command not found: $Name"
  }
}

function Get-DistEntry {
  return (Join-Path $McpRoot 'dist\index.js')
}

function Read-DotEnv([string]$Path) {
  $map = @{}
  if (-not (Test-Path -LiteralPath $Path)) { return $map }
  Get-Content -LiteralPath $Path | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq '' -or $line.StartsWith('#')) { return }
    $idx = $line.IndexOf('=')
    if ($idx -lt 1) { return }
    $key = $line.Substring(0, $idx).Trim()
    $val = $line.Substring($idx + 1).Trim()
    if (($val.StartsWith('"') -and $val.EndsWith('"')) -or ($val.StartsWith("'") -and $val.EndsWith("'"))) {
      $val = $val.Substring(1, $val.Length - 2)
    }
    $map[$key] = $val
  }
  return $map
}

function Test-McpInstalled {
  $dist = Get-DistEntry
  if (Test-Path -LiteralPath $dist) { return $true }
  if (-not (Test-Path -LiteralPath $CursorMcpPath)) { return $false }
  try {
    # Node parses case-sensitive JSON keys; ConvertFrom-Json fails on NO_PROXY/no_proxy.
    $flag = & node -e "const fs=require('fs');const p=process.argv[1];const n=process.argv[2];try{const c=JSON.parse(fs.readFileSync(p,'utf8'));process.stdout.write(c&&c.mcpServers&&Object.prototype.hasOwnProperty.call(c.mcpServers,n)?'1':'0')}catch{process.stdout.write('0')}" $CursorMcpPath $ServerName
    return ($flag -eq '1')
  } catch {
    return $false
  }
}

function Get-ExistingMcpServerEnv {
  if (-not (Test-Path -LiteralPath $CursorMcpPath)) { return @{} }
  # AGENT_BRAIN_* keys are unique ∴ ConvertFrom-Json is safe on this slice only.
  $json = & node -e "const fs=require('fs');const p=process.argv[1];const n=process.argv[2];let e={};try{const c=JSON.parse(fs.readFileSync(p,'utf8'));if(c&&c.mcpServers&&c.mcpServers[n]&&c.mcpServers[n].env)e=c.mcpServers[n].env}catch{}process.stdout.write(JSON.stringify(e))" $CursorMcpPath $ServerName
  if ([string]::IsNullOrWhiteSpace($json) -or $json -eq '{}') { return @{} }
  $obj = $json | ConvertFrom-Json
  $map = @{}
  foreach ($p in $obj.PSObject.Properties) { $map[$p.Name] = [string]$p.Value }
  return $map
}

function Write-Utf8NoBom([string]$Path, [string]$Content) {
  $utf8 = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($Path, $Content, $utf8)
}

function Write-CursorMcpServer([string]$EntryJson) {
  $entryPath = [System.IO.Path]::GetTempFileName()
  try {
    # Keep entry JSON verbatim (avoid PS 5.1 ConvertTo-Json flattening args:[x] → "x").
    Write-Utf8NoBom $entryPath $EntryJson
    & node -e "const fs=require('fs');const mcpPath=process.argv[1];const name=process.argv[2];const entry=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));let cfg={mcpServers:{}};if(fs.existsSync(mcpPath)){cfg=JSON.parse(fs.readFileSync(mcpPath,'utf8'));if(!cfg.mcpServers)cfg.mcpServers={}}cfg.mcpServers[name]=entry;fs.writeFileSync(mcpPath,JSON.stringify(cfg,null,2)+'\n','utf8')" $CursorMcpPath $ServerName $entryPath
    if ($LASTEXITCODE -ne 0) { throw "Failed to write Cursor MCP config ($LASTEXITCODE)" }
  } finally {
    Remove-Item -LiteralPath $entryPath -Force -ErrorAction SilentlyContinue
  }
}

function Ensure-EnvFile {
  $envPath = Join-Path $McpRoot '.env'
  $example = Join-Path $McpRoot '.env.example'
  if (Test-Path -LiteralPath $envPath) { return }
  if (-not (Test-Path -LiteralPath $example)) {
    throw "Missing .env and .env.example under $McpRoot"
  }
  Copy-Item -LiteralPath $example -Destination $envPath
  Write-Step "Created .env from .env.example"
}

function Invoke-NpmBuild {
  Push-Location $McpRoot
  try {
    Write-Step 'npm install'
    & npm install
    if ($LASTEXITCODE -ne 0) { throw "npm install failed ($LASTEXITCODE)" }
    Write-Step 'npm run build'
    & npm run build
    if ($LASTEXITCODE -ne 0) { throw "npm run build failed ($LASTEXITCODE)" }
  } finally {
    Pop-Location
  }
  $dist = Get-DistEntry
  if (-not (Test-Path -LiteralPath $dist)) {
    throw "Build succeeded but entry missing: $dist"
  }
}

function Get-McpServerEntry([hashtable]$ExistingEnv) {
  $distUnix = (Get-DistEntry) -replace '\\', '/'
  $dotenv = Read-DotEnv (Join-Path $McpRoot '.env')

  $pick = {
    param([string]$Key, [string]$Fallback)
    if ($ExistingEnv -and $ExistingEnv.ContainsKey($Key) -and $ExistingEnv[$Key]) {
      return [string]$ExistingEnv[$Key]
    }
    if ($dotenv[$Key]) { return [string]$dotenv[$Key] }
    return $Fallback
  }

  # Build JSON via Node so single-element args stays an array (PS 5.1 ConvertTo-Json flattens it).
  $envObj = [ordered]@{
    AGENT_BRAIN_API_URL  = & $pick 'AGENT_BRAIN_API_URL' 'http://127.0.0.1:8220'
    AGENT_BRAIN_USERNAME = & $pick 'AGENT_BRAIN_USERNAME' 'armin'
    AGENT_BRAIN_PASSWORD = & $pick 'AGENT_BRAIN_PASSWORD' 'dopadopa123'
  }
  $buildPath = [System.IO.Path]::GetTempFileName()
  try {
    $buildJson = (@{ dist = $distUnix; env = $envObj } | ConvertTo-Json -Depth 5 -Compress)
    Write-Utf8NoBom $buildPath $buildJson
    $entryJson = & node -e "const fs=require('fs');const i=JSON.parse(fs.readFileSync(process.argv[1],'utf8'));process.stdout.write(JSON.stringify({command:'node',args:[i.dist],env:i.env}))" $buildPath
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($entryJson)) {
      throw 'Failed to build MCP server entry JSON'
    }
    return $entryJson
  } finally {
    Remove-Item -LiteralPath $buildPath -Force -ErrorAction SilentlyContinue
  }
}

function Set-CursorMcpEntry {
  $cursorDir = Split-Path -Parent $CursorMcpPath
  if (-not (Test-Path -LiteralPath $cursorDir)) {
    New-Item -ItemType Directory -Path $cursorDir -Force | Out-Null
  }

  if (Test-Path -LiteralPath $CursorMcpPath) {
    $backup = "$CursorMcpPath.bak-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    Copy-Item -LiteralPath $CursorMcpPath -Destination $backup -Force
    Write-Step "Backed up Cursor MCP config → $backup"
  }

  $existingEnv = Get-ExistingMcpServerEnv
  $entryJson = Get-McpServerEntry $existingEnv
  Write-CursorMcpServer $entryJson
  Write-Step "Registered Cursor MCP '$ServerName' in $CursorMcpPath"
}

Assert-Command 'node'
Assert-Command 'npm'

$updating = Test-McpInstalled
if ($updating) {
  Write-Step "agent-brain-mcp already present → update"
} else {
  Write-Step "agent-brain-mcp not present → install"
}

Ensure-EnvFile
Invoke-NpmBuild
Set-CursorMcpEntry

if ($updating) {
  Write-Host "Done: updated '$ServerName' (restart Cursor MCP / reload window if needed)"
} else {
  Write-Host "Done: installed '$ServerName' (restart Cursor MCP / reload window if needed)"
}
