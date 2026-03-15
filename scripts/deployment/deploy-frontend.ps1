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
$RemoteAppDir = if ($env:REMOTE_APP_DIR) { $env:REMOTE_APP_DIR } else { "/home/anyadmin/app" }

# Standard SSH/SCP options - BatchMode=yes makes it non-interactive
$CommonSshArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-p", $RemotePort, "-i", $KeyFile)
$CommonScpArgs = @("-o", "BatchMode=yes", "-r", "-o", "StrictHostKeyChecking=no", "-P", $RemotePort, "-i", $KeyFile)

Write-Host "Starting Frontend Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 1. Build CSS
Write-Host "[1/5] Building Frontend CSS..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    npm run build:frontend:css
    if ($LASTEXITCODE -ne 0) { throw "CSS build failed" }
    Write-Host "CSS build successful." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 2. Stop Remote Frontend
Write-Host "[2/5] Stopping Remote Frontend..." -ForegroundColor Yellow
try {
    # Assuming the app is named 'app.js' or started by node
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "pkill -f 'node app.js' || true"
    Write-Host "Remote frontend stopped (if running)." -ForegroundColor Green
} catch {
    Write-Warning "Failed to stop frontend or connection issue: $_"
}

# 3. Upload Frontend files
Write-Host "[3/5] Uploading Frontend files..." -ForegroundColor Yellow
try {
    # Create remote directory
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "mkdir -p $RemoteAppDir && chown anyadmin:anyadmin $RemoteAppDir"
    
    # Upload files using scp (excluding node_modules)
    $TempStaging = "$ProjectRoot\tmp_frontend_staging"
    if (Test-Path $TempStaging) { Remove-Item -Recurse -Force $TempStaging }
    New-Item -ItemType Directory -Path $TempStaging | Out-Null
    
    # Copy frontend files excluding node_modules
    Copy-Item -Path "$FrontendDir\*" -Destination $TempStaging -Recurse -Exclude "node_modules"
    
    # Upload
    scp @CommonScpArgs "$TempStaging\*" "$RemoteUser@$RemoteHost`:$RemoteAppDir/"

    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "chown anyadmin:anyadmin $RemoteAppDir"
    
    if ($LASTEXITCODE -ne 0) { throw "SCP failed" }
    Write-Host "Upload successful." -ForegroundColor Green
}
finally {
    if (Test-Path $TempStaging) { Remove-Item -Recurse -Force $TempStaging }
}

# 4. Install Dependencies on Remote
Write-Host "[4/5] Installing Dependencies on Remote..." -ForegroundColor Yellow
try {
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "runuser -l anyadmin -c 'cd $RemoteAppDir && npm install --production'"
    if ($LASTEXITCODE -ne 0) { throw "Remote npm install failed" }
    Write-Host "Dependencies installed successfully." -ForegroundColor Green
} catch {
    throw "Failed to install dependencies: $_"
}

# 5. Start Remote Frontend
Write-Host "[5/5] Starting Remote Frontend..." -ForegroundColor Yellow
$StartCmd = "runuser -l anyadmin -c 'cd $RemoteAppDir && (nohup node app.js > /home/anyadmin/logs/frontend.log 2>&1 < /dev/null &)'"
try {
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" $StartCmd
    if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
    Write-Host "Frontend started successfully." -ForegroundColor Green
} catch {
    throw "Failed to start frontend: $_"
}

Write-Host "Deployment Complete!" -ForegroundColor Cyan
