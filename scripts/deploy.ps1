# DownCollect Deploy Script
# Usage: .\scripts\deploy.ps1 [-RegistryName <name>] [-ResourceGroup <rg>] [-AppName <app>]
#
# Builds the Docker image via ACR Build, tags it with semantic version + build number,
# retags to short version and latest, then updates the Container App.

param(
    [string]$RegistryName = "downcollectacr",
    [string]$ResourceGroup = "downcollect-rg",
    [string]$AppName = "downcollect"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

# Read semantic version
$Version = (Get-Content "$Root\VERSION" -Raw).Trim()
if (-not $Version) { Write-Error "VERSION file is empty"; exit 1 }

# Sync package.json version from VERSION
$PkgJson = "$Root\src\frontend\package.json"
$Pkg = Get-Content $PkgJson -Raw | ConvertFrom-Json
if ($Pkg.version -ne $Version) {
    Write-Host "Syncing package.json version: $($Pkg.version) -> $Version" -ForegroundColor Yellow
    $Pkg.version = $Version
    $Pkg | ConvertTo-Json -Depth 10 | Set-Content $PkgJson -Encoding UTF8
}

# Build number: YYYYMMDD-HHmm (UTC)
$BuildNum = (Get-Date).ToUniversalTime().ToString("yyyyMMdd-HHmm")
$FullTag = "$Version.$BuildNum"
$ShortTag = $Version
$ImageRepo = "downcollect"

Write-Host "=== Deploying DownCollect ===" -ForegroundColor Cyan
Write-Host "  Version:    $Version"
Write-Host "  Full tag:   $FullTag"
Write-Host "  Registry:   $RegistryName"
Write-Host "  App:        $AppName"
Write-Host ""

# Step 1: Build and push image via ACR Build
Write-Host "--- Building image via ACR Build ---" -ForegroundColor Yellow
Push-Location $Root
try {
    az acr build `
        --registry $RegistryName `
        --image "${ImageRepo}:${FullTag}" `
        --file deploy/Dockerfile .
} finally {
    Pop-Location
}

# Step 2: Retag to short version and latest
Write-Host "--- Retagging to ${ShortTag} and latest ---" -ForegroundColor Yellow
az acr import `
    --name $RegistryName `
    --source "${RegistryName}.azurecr.io/${ImageRepo}:${FullTag}" `
    --image "${ImageRepo}:${ShortTag}" `
    --force

az acr import `
    --name $RegistryName `
    --source "${RegistryName}.azurecr.io/${ImageRepo}:${FullTag}" `
    --image "${ImageRepo}:latest" `
    --force

# Step 3: Update Container App
Write-Host "--- Updating Container App ---" -ForegroundColor Yellow
az containerapp update `
    --name $AppName `
    --resource-group $ResourceGroup `
    --image "${RegistryName}.azurecr.io/${ImageRepo}:${FullTag}"

# Step 4: Verify
Write-Host ""
Write-Host "=== Deployment complete ===" -ForegroundColor Green
Write-Host "  Image: ${RegistryName}.azurecr.io/${ImageRepo}:${FullTag}"
Write-Host "  Tags:  ${FullTag}, ${ShortTag}, latest"

$Fqdn = az containerapp show --name $AppName --resource-group $ResourceGroup --query "properties.configuration.ingress.fqdn" --output tsv
Write-Host "  URL:   https://$Fqdn"
