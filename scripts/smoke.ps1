$ErrorActionPreference = "Stop"

$frontendPort = if ($env:FRONTEND_PORT) { $env:FRONTEND_PORT } else { "5173" }
$backendPort = if ($env:BACKEND_PORT) { $env:BACKEND_PORT } else { "8080" }
$today = Get-Date -Format "yyyy-MM-dd"

$health = Invoke-RestMethod "http://localhost:$backendPort/api/health"
if ($health.status -ne "ok") {
    throw "El backend no respondió correctamente."
}

$dashboard = Invoke-RestMethod "http://localhost:$backendPort/api/dashboard?date=$today"
if ($dashboard.tasks.Count -lt 4) {
    throw "El checklist diario no fue creado correctamente."
}

$frontend = Invoke-WebRequest "http://localhost:$frontendPort" -UseBasicParsing
if ($frontend.StatusCode -ne 200 -or $frontend.Content -notmatch "CalorieTrack") {
    throw "El frontend no respondió correctamente."
}

Write-Host "Smoke test correcto: db, backend y frontend están conectados." -ForegroundColor Green

