$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    docker compose down
    Write-Host "CalorieTrack se detuvo. Los datos de PostgreSQL se conservaron." -ForegroundColor Yellow
} finally {
    Pop-Location
}

