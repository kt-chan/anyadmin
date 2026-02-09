# Deploy Agent Script
$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path "$PSScriptRoot\.."
$BackendDir = "$ProjectRoot\backend"
$KeyFile = "$BackendDir\keys\id_rsa"
$RemoteUser = "root"
$RemoteHost = "172.20.0.10"
$RemoteIP = "172.20.0.10"
$RemoteBinDir = "/home/anyadmin/bin"
$AgentName = "anyadmin-agent"

Write-Host "Starting Agent Deployment..." -ForegroundColor Cyan

# 1. Compile Agent
Write-Host "[1/4] Compiling Agent for Linux/AMD64..." -ForegroundColor Yellow
Push-Location $BackendDir
try {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -o "./dist/$AgentName" ./cmd/agent/main.go
    if ($LASTEXITCODE -ne 0) { throw "Compilation failed" }
    Write-Host "Compilation successful." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 2. Stop Remote Agent
Write-Host "[2/4] Stopping Remote Agent..." -ForegroundColor Yellow
try {
    ssh -o StrictHostKeyChecking=no -i $KeyFile "$RemoteUser@$RemoteHost" "pkill -9 $AgentName || true"
    Write-Host "Remote agent stopped (if running)." -ForegroundColor Green
} catch {
    Write-Warning "Failed to stop agent or connection issue: $_"
}

# 3. Upload Agent
Write-Host "[3/4] Uploading Agent Binary..." -ForegroundColor Yellow
try {
    scp -o StrictHostKeyChecking=no -i $KeyFile "$BackendDir\dist\$AgentName" "$RemoteUser@$RemoteHost`:$RemoteBinDir/$AgentName"
    if ($LASTEXITCODE -ne 0) { throw "SCP failed" }
    Write-Host "Upload successful." -ForegroundColor Green
} catch {
    throw "Upload failed: $_"
}

# 4. Start Remote Agent
Write-Host "[4/4] Starting Remote Agent..." -ForegroundColor Yellow
$StartCmd = "chmod +x $RemoteBinDir/$AgentName && runuser -l anyadmin -c 'cd $RemoteBinDir && (nohup ./$AgentName -config config.json -log /home/anyadmin/logs/agent.log > /home/anyadmin/logs/agent.log 2>&1 < /dev/null &)'"
try {
    ssh -o StrictHostKeyChecking=no -i $KeyFile "$RemoteUser@$RemoteHost" $StartCmd
    if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
    Write-Host "Agent started successfully." -ForegroundColor Green
} catch {
    throw "Failed to start agent: $_"
}

Write-Host "Deployment Complete!" -ForegroundColor Cyan
