# DownCollect Build Scripts
# Usage: .\scripts\build.ps1 [target]
# Targets: frontend, backend, all, clean

param(
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

function Build-Frontend {
    Write-Host "=== Building frontend ===" -ForegroundColor Cyan
    Push-Location "$Root\src\frontend"
    try {
        if (-not (Test-Path node_modules)) {
            Write-Host "Installing npm dependencies..."
            npm install
        }
        npx vite build --outDir "$Root\dist\frontend" --emptyOutDir
    } finally {
        Pop-Location
    }
}

function Build-Backend {
    Write-Host "=== Building backend ===" -ForegroundColor Cyan
    Push-Location "$Root\src\backend"
    try {
        $outDir = "$Root\dist\backend"
        if (-not (Test-Path $outDir)) {
            New-Item -ItemType Directory -Path $outDir | Out-Null
        }
        go build -o "$outDir\server.exe" ./cmd/server/
        Write-Host "Binary: $outDir\server.exe"

        # Copy frontend build next to server binary as "public/"
        $frontendDist = "$Root\dist\frontend"
        $publicDir = "$outDir\public"
        if (Test-Path $frontendDist) {
            if (Test-Path $publicDir) { Remove-Item -Recurse -Force $publicDir }
            Copy-Item -Recurse $frontendDist $publicDir
            Write-Host "Frontend copied to $publicDir"
        } else {
            Write-Host "Warning: Frontend not built yet. Run 'build.ps1 frontend' first for full package." -ForegroundColor Yellow
        }
    } finally {
        Pop-Location
    }
}

function Test-Backend {
    Write-Host "=== Testing backend ===" -ForegroundColor Cyan
    Push-Location "$Root\src\backend"
    try {
        go test ./internal/... -count=1 -timeout 30s
    } finally {
        Pop-Location
    }
}

function Clean-Build {
    Write-Host "=== Cleaning build output ===" -ForegroundColor Cyan
    $dirs = @("$Root\dist", "$Root\src\frontend\dist", "$Root\src\backend\dist")
    foreach ($d in $dirs) {
        if (Test-Path $d) { Remove-Item -Recurse -Force $d; Write-Host "Removed $d" }
    }
    Write-Host "Cleaned."
}

switch ($Target) {
    "frontend" { Build-Frontend }
    "backend"  { Build-Backend }
    "test"     { Test-Backend }
    "all"      { Build-Frontend; Build-Backend }
    "clean"    { Clean-Build }
    default    { Write-Host "Unknown target: $Target. Use: frontend, backend, test, all, clean" }
}
