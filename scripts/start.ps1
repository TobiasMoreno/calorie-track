$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    docker compose up --build --detach
    docker compose ps
    Write-Host ""
    Write-Host "CalorieTrack está disponible en http://localhost:5173" -ForegroundColor Green
    Write-Host "API: http://localhost:8080/api/health"
} finally {
    Pop-Location
}

