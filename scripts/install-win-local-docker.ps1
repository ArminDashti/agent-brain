#Requires -Version 5.1
<#
.SYNOPSIS
  Install or update the local Docker stack for agent-brain (Windows).

.DESCRIPTION
  Compose project: agent-brain
  Services: agent-brain-api (host 5090), agent-brain-webui (host 5091), plus Postgres + Qdrant.
  Fresh install when the stack is absent; otherwise rebuild images and recreate containers.
  Named volumes are kept so Postgres/Qdrant data is not wiped on update.

.EXAMPLE
  .\install-win-local-docker.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$StackName = 'agent-brain'
$ComposeFile = Join-Path $PSScriptRoot 'docker-compose.yml'
$ApiPort = 5090
$WebUiPort = 5091

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

$updating = Test-StackPresent
if ($updating) {
  Write-Step ("Stack '{0}' present -> update (rebuild images, keep volumes)" -f $StackName)
} else {
  Write-Step "Stack '$StackName' absent → install"
}

Write-Step 'docker compose up -d --build'
Invoke-Compose @('up', '-d', '--build')

Write-Host "Done: stack '$StackName' running"
Write-Host "  API    http://127.0.0.1:$ApiPort"
Write-Host "  WebUI  http://127.0.0.1:$WebUiPort"
if ($updating) {
  Write-Host '  Data volumes preserved (Postgres + Qdrant).'
}
