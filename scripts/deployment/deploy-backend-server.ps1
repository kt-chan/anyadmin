# Deploy Backend Server Script
$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path "$PSScriptRoot\..\.."
$BackendDir = "$ProjectRoot\backend"
$KeyFile = "$BackendDir\keys\id_rsa"
$ServerName = "anyadmin-server"

# Load environment variables
. "$PSScriptRoot\utils.ps1"
Load-Env -Path "$ProjectRoot\.env"

$RemoteUser = $env:REMOTE_USER -or "root"
$RemoteHost = $env:REMOTE_HOST -or "172.25.208.100"
$RemotePort = $env:REMOTE_SSH_PORT -or "22"
$RemoteBinDir = $env:REMOTE_BIN_DIR -or "/home/anyadmin/bin"

# Standard SSH/SCP options
$SshOpts = "-o StrictHostKeyChecking=no -p $RemotePort -i $KeyFile"
$ScpOpts = "-o StrictHostKeyChecking=no -P $RemotePort -i $KeyFile"

Write-Host "Starting Backend Server Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 1. Compile Server
Write-Host "[1/4] Compiling Server for Linux/AMD64..." -ForegroundColor Yellow
Push-Location $BackendDir
try {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -o "./dist/linux_amd64/$ServerName" ./cmd/server/main.go
    if ($LASTEXITCODE -ne 0) { throw "Compilation failed" }
    Write-Host "Compilation successful." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 2. Stop Remote Server
Write-Host "[2/4] Stopping Remote Server..." -ForegroundColor Yellow
try {
    ssh $SshOpts "$RemoteUser@$RemoteHost" "pkill -9 $ServerName || true"
    Write-Host "Remote server stopped (if running)." -ForegroundColor Green
} catch {
    Write-Warning "Failed to stop server or connection issue: $_"
}

# 3. Upload Server
Write-Host "[3/4] Uploading Server Binary..." -ForegroundColor Yellow
try {
    scp $ScpOpts "$BackendDir\dist\linux_amd64\$ServerName" "$RemoteUser@$RemoteHost`:$RemoteBinDir/$ServerName"
    if ($LASTEXITCODE -ne 0) { throw "SCP failed" }
    Write-Host "Upload successful." -ForegroundColor Green
} catch {
    throw "Upload failed: $_"
}

# 4. Start Remote Server
Write-Host "[4/4] Starting Remote Server..." -ForegroundColor Yellow
$StartCmd = "chmod +x $RemoteBinDir/$ServerName && runuser -l anyadmin -c 'cd $RemoteBinDir && (nohup ./$ServerName > /home/anyadmin/logs/server.log 2>&1 < /dev/null &)'"
try {
    ssh $SshOpts "$RemoteUser@$RemoteHost" $StartCmd
    if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
    Write-Host "Server started successfully." -ForegroundColor Green
} catch {
    throw "Failed to start server: $_"
}

Write-Host "Deployment Complete!" -ForegroundColor Cyan
