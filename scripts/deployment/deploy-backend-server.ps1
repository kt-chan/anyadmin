# Deploy Backend Server Script
$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path "$PSScriptRoot\..\.."
$BackendDir = "$ProjectRoot\backend"
$KeyFile = "$BackendDir\keys\id_rsa"
$ServerName = "anyadmin-server"

# Load environment variables
. "$PSScriptRoot\utils.ps1"
Load-Env -Path "$ProjectRoot\.env"

$RemoteUser = if ($env:REMOTE_USER) { $env:REMOTE_USER } else { "root" }
$RemoteHost = if ($env:REMOTE_HOST) { $env:REMOTE_HOST } else { "172.25.208.100" }
$RemotePort = if ($env:REMOTE_SSH_PORT) { $env:REMOTE_SSH_PORT } else { "22" }
$RemoteBinDir = if ($env:REMOTE_BIN_DIR) { $env:REMOTE_BIN_DIR } else { "/home/anyadmin/app" }
$RemoteSrcDir = if ($env:REMOTE_SRC_DIR) { $env:REMOTE_SRC_DIR } else { "/home/anyadmin/src" }

# Standard SSH/SCP options - BatchMode=yes makes it non-interactive
$CommonSshArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-p", $RemotePort, "-i", $KeyFile)
$CommonScpArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-P", $RemotePort, "-i", $KeyFile)

Write-Host "Starting Backend Server Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 1. Sync Project Source to Remote
Write-Host "[1/5] Syncing Project Source to Remote..." -ForegroundColor Yellow
Sync-RemoteSource `
    -LocalPath $ProjectRoot `
    -RemotePath $RemoteSrcDir `
    -ArchiveName "project_src" `
    -RemoteUser $RemoteUser `
    -RemoteHost $RemoteHost `
    -CommonSshArgs $CommonSshArgs `
    -CommonScpArgs $CommonScpArgs

# 2. Build Server on Remote Host
Write-Host "[2/5] Building Server on Remote Host..." -ForegroundColor Yellow
$BuildCmd = "source /etc/profile.d/go.sh 2>/dev/null || true; mkdir -p $RemoteBinDir && cd $RemoteSrcDir/backend && go build -o $RemoteBinDir/$ServerName ./cmd/server/main.go"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$BuildCmd"
if ($LASTEXITCODE -ne 0) { throw "Remote compilation failed" }
Write-Host "Remote compilation successful." -ForegroundColor Green

# 3. Stop Remote Server
Write-Host "[3/5] Stopping Remote Server..." -ForegroundColor Yellow
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "pkill -9 $ServerName || true"
Write-Host "Remote server stopped (if running)." -ForegroundColor Green

# 4. Upload Config Files
Write-Host "[4/5] Uploading Configuration Files..." -ForegroundColor Yellow
scp @CommonScpArgs "$BackendDir\config.yaml" "$RemoteUser@$RemoteHost`:$RemoteBinDir/config.yaml"
scp @CommonScpArgs "$BackendDir\data.json" "$RemoteUser@$RemoteHost`:$RemoteBinDir/data.json"
Write-Host "Configs uploaded." -ForegroundColor Green

# 5. Start Remote Server
Write-Host "[5/5] Starting Remote Server..." -ForegroundColor Yellow
$StartCmd = "chmod +x $RemoteBinDir/$ServerName && runuser -l anyadmin -c 'cd $RemoteBinDir && (nohup ./$ServerName > /home/anyadmin/logs/server.log 2>&1 < /dev/null &)'"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$StartCmd"
if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
Write-Host "Server started successfully." -ForegroundColor Green

Write-Host "Deployment Complete!" -ForegroundColor Cyan
