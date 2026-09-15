#Requires -Version 5.1
<#
.SYNOPSIS
  Reinstall the local Docker stack for agent-brain (Windows).

.DESCRIPTION
  Runs remove-win-local-docker.ps1 then install-win-local-docker.ps1.
  This wipes containers, DB volumes, and images, then builds a fresh stack.

.EXAMPLE
  .\reinstall-win-local-docker.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$RemoveScript = Join-Path $PSScriptRoot 'remove-win-local-docker.ps1'
$InstallScript = Join-Path $PSScriptRoot 'install-win-local-docker.ps1'

if (-not (Test-Path -LiteralPath $RemoveScript)) {
  throw "Missing: $RemoveScript"
}
if (-not (Test-Path -LiteralPath $InstallScript)) {
  throw "Missing: $InstallScript"
}

Write-Host '==> Reinstall: remove then install'
& $RemoveScript
if ($LASTEXITCODE -ne 0) {
  throw "remove-win-local-docker.ps1 failed ($LASTEXITCODE)"
}

& $InstallScript
if ($LASTEXITCODE -ne 0) {
  throw "install-win-local-docker.ps1 failed ($LASTEXITCODE)"
}

Write-Host 'Done: reinstall complete.'
