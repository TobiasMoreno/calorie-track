$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    docker compose up --detach --wait db

    $postgresUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "calorietrack" }
    $postgresPassword = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "calorietrack_local" }
    $postgresPort = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { "5432" }
    $postgresDB = if ($env:POSTGRES_DB) { $env:POSTGRES_DB } else { "calorietrack" }
    $env:CALORIETRACK_TEST_DATABASE_URL = "postgres://$postgresUser`:$postgresPassword@localhost`:$postgresPort/$postgresDB`?sslmode=disable"

    Push-Location backend
    try {
        go test ./...
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }

    Push-Location frontend
    try {
        npm ci
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        npm run build
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }

    Write-Host "Tests de backend y build de frontend correctos." -ForegroundColor Green
} finally {
    Pop-Location
}

