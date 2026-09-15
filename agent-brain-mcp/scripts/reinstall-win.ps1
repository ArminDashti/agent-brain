#Requires -Version 5.1
<#
.SYNOPSIS
  Remove agent-brain-mcp completely, then install fresh.

.DESCRIPTION
  Runs remove-win.ps1 then install-win.ps1.

.EXAMPLE
  .\reinstall-win.ps1
#>
[CmdletBinding()]
param(
  [switch]$RemoveEnv
)

$ErrorActionPreference = 'Stop'

$RemoveScript = Join-Path $PSScriptRoot 'remove-win.ps1'
$InstallScript = Join-Path $PSScriptRoot 'install-win.ps1'

if (-not (Test-Path -LiteralPath $RemoveScript)) {
  throw "Missing $RemoveScript"
}
if (-not (Test-Path -LiteralPath $InstallScript)) {
  throw "Missing $InstallScript"
}

Write-Host '==> reinstall: remove completely then install'
if ($RemoveEnv) {
  & $RemoveScript -RemoveEnv
} else {
  & $RemoveScript
}
if (-not $?) {
  throw 'remove-win.ps1 failed'
}

Write-Host '==> reinstall: install'
& $InstallScript
if (-not $?) {
  throw 'install-win.ps1 failed'
}

Write-Host 'Done: reinstall complete'
