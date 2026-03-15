# Overall Deployment Script
$ErrorActionPreference = "Stop"

Write-Host "Starting Overall Deployment Process..." -ForegroundColor Magenta

# Get the path to the individual deployment scripts
$DeployDir = $PSScriptRoot
$AgentScript = Join-Path $DeployDir "deploy-agent.ps1"
$ServerScript = Join-Path $DeployDir "deploy-backend-server.ps1"
$FrontendScript = Join-Path $DeployDir "deploy-frontend.ps1"

# 1. Deploy Agent
Write-Host "`n>>> [1/3] Deploying Backend Agent..." -ForegroundColor Cyan
& $AgentScript

# 2. Deploy Backend Server
Write-Host "`n>>> [2/3] Deploying Backend Server..." -ForegroundColor Cyan
& $ServerScript

# 3. Deploy Frontend
Write-Host "`n>>> [3/3] Deploying Frontend..." -ForegroundColor Cyan
& $FrontendScript

Write-Host "`nOverall Deployment Complete!" -ForegroundColor Magenta
