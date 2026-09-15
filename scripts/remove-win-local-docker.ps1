#Requires -Version 5.1
<#
.SYNOPSIS
  Completely remove the local Docker stack for agent-brain (Windows).

.DESCRIPTION
  Stops and removes containers, named volumes (Postgres + Qdrant DB data),
  and project images for Compose project agent-brain.

.EXAMPLE
  .\remove-win-local-docker.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$StackName = 'agent-brain'
$ComposeFile = Join-Path $PSScriptRoot 'docker-compose.yml'
$ProjectImages = @('agent-brain-api:local', 'agent-brain-webui:local')

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

if (-not (Test-StackPresent)) {
  Write-Step "Stack '$StackName' not present - cleaning leftover images if any"
} else {
  Write-Step ("Removing stack '{0}' containers, volumes, and images" -f $StackName)
}

# --rmi local removes images built by this compose project; -v drops named volumes (DB)
Write-Step 'docker compose down -v --rmi local --remove-orphans'
Invoke-Compose @('down', '-v', '--rmi', 'local', '--remove-orphans')

foreach ($img in $ProjectImages) {
  $found = & docker images -q $img
  if ($found) {
    Write-Step "Removing image $img"
    & docker rmi -f $img 1>$null 2>$null
  }
}

Write-Host ("Done: stack '{0}' removed (containers, DB volumes, images)." -f $StackName)
