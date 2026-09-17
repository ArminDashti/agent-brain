#Requires -Version 5.1
<#
.SYNOPSIS
  Install or update the local Docker stack for agent-brain (Windows).

.DESCRIPTION
  Compose project: agent-brain
  Services: agent-brain-api (host 5090), agent-brain-webui (internal + pc-armin-local), Qdrant + SQLite.
  Routes http://agent-brain.local via nginx-gateway (port 80 already in use by gateway).
  Ensures hosts entry for agent-brain.local and nginx-local conf.

.EXAMPLE
  .\install-win-local-docker.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$StackName = 'agent-brain'
$ComposeFile = Join-Path $PSScriptRoot 'docker-compose.yml'
$ApiPort = 5090
$HostsName = 'agent-brain.local'
$HostsLine = "127.0.0.1 $HostsName"
$NginxConfDir = 'C:\Users\armin\GitHub\nginx-local\conf'
$NginxConfName = 'agent-brain.conf'

function Write-Step([string]$Message) {
  Write-Host "==> $Message"
}

function Assert-Command([string]$Name) {
  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command not found: $Name"
  }
}

function Assert-DockerReady {
  Assert-Command 'docker'
  & docker info 1>$null 2>$null
  if ($LASTEXITCODE -ne 0) {
    throw 'Docker daemon is not reachable. Start Docker Desktop and retry.'
  }
  & docker compose version 1>$null 2>$null
  if ($LASTEXITCODE -ne 0) {
    throw 'Docker Compose plugin is required (docker compose).'
  }
}

function Ensure-HostsEntry {
  $hostsPath = Join-Path $env:SystemRoot 'System32\drivers\etc\hosts'
  $content = Get-Content -LiteralPath $hostsPath -ErrorAction Stop
  $exists = $false
  foreach ($line in $content) {
    if ($line -match '^\s*127\.0\.0\.1\s+agent-brain\.local(\s|$)') {
      $exists = $true
      break
    }
  }
  if ($exists) {
    Write-Step "Hosts already has $HostsName"
    return
  }
  Write-Step "Adding hosts entry: $HostsLine"
  try {
    Add-Content -LiteralPath $hostsPath -Value $HostsLine -ErrorAction Stop
  } catch {
    Write-Host "WARN: could not write hosts file (run as Administrator): $($_.Exception.Message)"
    Write-Host "Add manually: $HostsLine"
  }
}

function Ensure-NginxGatewayConf {
  $dest = Join-Path $NginxConfDir $NginxConfName
  $src = Join-Path $PSScriptRoot $NginxConfName
  if (-not (Test-Path -LiteralPath $NginxConfDir)) {
    Write-Host "WARN: nginx-local conf dir missing: $NginxConfDir"
    return
  }
  if (Test-Path -LiteralPath $src) {
    Write-Step "Installing nginx gateway conf -> $dest"
    Copy-Item -LiteralPath $src -Destination $dest -Force
  } elseif (-not (Test-Path -LiteralPath $dest)) {
    Write-Host "WARN: missing $NginxConfName in scripts/ and nginx-local"
    return
  }
  $gw = & docker ps -q --filter 'name=nginx-gateway'
  if ($gw) {
    Write-Step 'Reloading nginx-gateway'
    & docker exec nginx-gateway nginx -s reload
    if ($LASTEXITCODE -ne 0) {
      Write-Host 'WARN: nginx reload failed; restart nginx-gateway manually'
    }
  } else {
    Write-Host 'WARN: nginx-gateway container not running'
  }
}

function Test-StackPresent {
  $ids = & docker ps -a -q --filter "label=com.docker.compose.project=$StackName"
  if ($LASTEXITCODE -ne 0) {
    throw 'Failed to query Docker containers.'
  }
  return [bool]($ids -and @($ids).Count -gt 0)
}

function Invoke-Compose {
  param([Parameter(Mandatory)][string[]]$ComposeArgs)
  & docker compose -p $StackName -f $ComposeFile @ComposeArgs
  if ($LASTEXITCODE -ne 0) {
    throw "docker compose failed ($LASTEXITCODE): $($ComposeArgs -join ' ')"
  }
}

Assert-DockerReady

if (-not (Test-Path -LiteralPath $ComposeFile)) {
  throw "Compose file missing: $ComposeFile"
}

$extNet = & docker network ls -q --filter 'name=^pc-armin-local$'
if (-not $extNet) {
  throw 'Docker network pc-armin-local is missing (required for agent-brain.local via nginx-gateway).'
}

Ensure-HostsEntry
Ensure-NginxGatewayConf

$updating = Test-StackPresent
if ($updating) {
  Write-Step ("Stack '{0}' present -> update (rebuild images, keep volumes)" -f $StackName)
} else {
  Write-Step "Stack '$StackName' absent → install"
}

Write-Step 'docker compose up -d --build'
Invoke-Compose @('up', '-d', '--build')

Write-Host "Done: stack '$StackName' running"
Write-Host "  WebUI  http://$HostsName"
Write-Host "  API    http://127.0.0.1:$ApiPort"
if ($updating) {
  Write-Host '  Data volumes preserved (SQLite + Qdrant).'
} else {
  Write-Host '  Fresh SQLite volume (no Postgres).'
}
