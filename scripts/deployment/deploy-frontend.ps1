# Deploy Frontend Script
$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path "$PSScriptRoot\..\.."
$FrontendDir = "$ProjectRoot\frontend"
$BackendDir = "$ProjectRoot\backend"
$KeyFile = "$BackendDir\keys\id_rsa"

# Load environment variables
. "$PSScriptRoot\utils.ps1"
Load-Env -Path "$ProjectRoot\.env"

$RemoteUser = if ($env:REMOTE_USER) { $env:REMOTE_USER } else { "root" }
$RemoteHost = if ($env:REMOTE_HOST) { $env:REMOTE_HOST } else { "172.25.208.100" }
$RemotePort = if ($env:REMOTE_SSH_PORT) { $env:REMOTE_SSH_PORT } else { "22" }
$RemoteAppDir = if ($env:REMOTE_APP_DIR) { $env:REMOTE_APP_DIR } else { "/home/anyadmin/app/frontend" }
$RemoteSrcDir = if ($env:REMOTE_SRC_DIR) { $env:REMOTE_SRC_DIR } else { "/home/anyadmin/src" }

# Standard SSH/SCP options - BatchMode=yes makes it non-interactive
$CommonSshArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-p", $RemotePort, "-i", $KeyFile)
$CommonScpArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-P", $RemotePort, "-i", $KeyFile)

Write-Host "Starting Frontend Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 0. Ensure Node.js is up to date
Ensure-RemoteNodeJS -RemoteUser $RemoteUser -RemoteHost $RemoteHost -CommonSshArgs $CommonSshArgs

# 1. Sync Project Source to Remote (Including root package.json for tailwind)
Write-Host "[1/5] Syncing Project Source to Remote..." -ForegroundColor Yellow
Sync-RemoteSource `
    -LocalPath $ProjectRoot `
    -RemotePath $RemoteSrcDir `
    -ArchiveName "project_src" `
    -RemoteUser $RemoteUser `
    -RemoteHost $RemoteHost `
    -CommonSshArgs $CommonSshArgs `
    -CommonScpArgs $CommonScpArgs

# 2. Stop Remote Frontend
Write-Host "[2/5] Stopping Remote Frontend..." -ForegroundColor Yellow
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "pkill -f 'node app.js' || true"
Write-Host "Remote frontend stopped (if running)." -ForegroundColor Green

# 3. Build Frontend on Remote (Dependencies and CSS)
Write-Host "[3/5] Building Frontend on Remote (npm install and CSS build)..." -ForegroundColor Yellow
# Run install in root for tailwind and in frontend for app dependencies
$BuildCmd = "cd $RemoteSrcDir && npm install && cd frontend && npm install --production && cd .. && npm run build:frontend:css"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$BuildCmd"
Write-Host "Remote build completed." -ForegroundColor Green

# 4. Prepare App Directory
Write-Host "[4/5] Syncing to App Directory..." -ForegroundColor Yellow
$PrepareCmd = "mkdir -p $RemoteAppDir && cp -r $RemoteSrcDir/frontend/* $RemoteAppDir/ && cp $RemoteSrcDir/.env $RemoteAppDir/ && chown anyadmin:anyadmin -R $RemoteAppDir"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$PrepareCmd"
Write-Host "App directory prepared." -ForegroundColor Green

# 5. Start Remote Frontend
Write-Host "[5/5] Starting Remote Frontend..." -ForegroundColor Yellow
$StartCmd = "runuser -l anyadmin -c 'cd $RemoteAppDir && (nohup node app.js > /home/anyadmin/logs/frontend.log 2>&1 < /dev/null &)'"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$StartCmd"
if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
Write-Host "Frontend started successfully." -ForegroundColor Green

Write-Host "Deployment Complete!" -ForegroundColor Cyan
