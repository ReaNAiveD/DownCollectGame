---
description: "Deploy DownCollect to Azure Container App. Use when asked to deploy, release, or update the production environment."
---

# Deploy to Azure Container App

## Prerequisites

- Azure CLI logged in (`az login`)
- The user should provide the target resource group, container registry, and container app name.
- If the user doesn't provide them, the agent should ask user to input or confirm the default values:
  - Resource Group: `downcollect-rg`
  - Container Registry: `downcollectacr`
  - Container App Name: `downcollect`

## Versioning Scheme

- **VERSION file** at repo root contains the semantic version (e.g. `0.1.0`).
- **Image tag format**: `{semver}.{YYYYMMDD-HHmm}` — e.g. `0.1.0.20260316-0830`
  - The build number is UTC timestamp to the minute, ensuring uniqueness and chronological sorting.
- **Retag**: After building, retag to short version (`0.1.0`) and `latest` using `az acr import --force`.

## Quick Deploy (use the script)

```powershell
.\scripts\deploy.ps1
```

This handles everything: build, tag, retag, update container app.

## Manual Steps (if needed)

### 1. Build and push image via ACR Build

**IMPORTANT**: Use `az acr build` directly — it builds AND pushes in one command on Azure's servers. Do NOT pipe to docker push or use local docker build.

```powershell
$Version = (Get-Content VERSION -Raw).Trim()
$BuildNum = (Get-Date).ToUniversalTime().ToString("yyyyMMdd-HHmm")
$FullTag = "$Version.$BuildNum"

az acr build --registry downcollectacr --image "downcollect:$FullTag" --file deploy/Dockerfile .
```

### 2. Retag to short version and latest

Use `az acr import` with `--force` to create additional tags pointing to the same image:

```powershell
az acr import --name downcollectacr --source "downcollectacr.azurecr.io/downcollect:$FullTag" --image "downcollect:$Version" --force
az acr import --name downcollectacr --source "downcollectacr.azurecr.io/downcollect:$FullTag" --image "downcollect:latest" --force
```

### 3. Update Container App

Always deploy with the **full tag** (not `latest`) so the revision is traceable:

```powershell
az containerapp update --name downcollect --resource-group downcollect-rg --image "downcollectacr.azurecr.io/downcollect:$FullTag"
```

### 4. Verify

```powershell
az containerapp show --name downcollect --resource-group downcollect-rg --query "properties.configuration.ingress.fqdn" --output tsv
```

## Version Bump

When releasing a new version, edit `VERSION` file:
- **Patch** (bug fixes): `0.1.0` → `0.1.1`
- **Minor** (new features): `0.1.0` → `0.2.0`
- **Major** (breaking changes): `0.1.0` → `1.0.0`

## Rollback

To rollback, redeploy a previous image tag:

```powershell
az containerapp update --name downcollect --resource-group downcollect-rg --image "downcollectacr.azurecr.io/downcollect:0.1.0.20260315-0730"
```

List available tags:
```powershell
az acr repository show-tags --name downcollectacr --repository downcollect --orderby time_desc --output table
```

## Dockerfile Notes

- The Dockerfile uses `node node_modules/vite/bin/vite.js build` instead of `npx vite build` because ACR Build's restricted Docker environment gives `Permission denied` on `node_modules/.bin` symlinks.
- Go version in Dockerfile must match `go.mod`.

## Agent Limitations

- **NEVER** pipe `az acr build` output through `Select-Object`, `Select-String`, `Out-String`, or any other PowerShell filter. The command streams build logs in real time and piping can truncate output, hide errors, or cause the command to appear to hang. Always run `az acr build` as a bare command with no output redirection.
