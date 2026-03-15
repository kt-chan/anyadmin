# Deploy Agent Script
$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path "$PSScriptRoot\..\.."
$BackendDir = "$ProjectRoot\backend"
$KeyFile = "$BackendDir\keys\id_rsa"
$AgentName = "anyadmin-agent"

# Load environment variables
. "$PSScriptRoot\utils.ps1"
Load-Env -Path "$ProjectRoot\.env"

$RemoteUser = if ($env:REMOTE_USER) { $env:REMOTE_USER } else { "root" }
$RemoteHost = if ($env:REMOTE_HOST) { $env:REMOTE_HOST } else { "172.25.208.100" }
$RemotePort = if ($env:REMOTE_SSH_PORT) { $env:REMOTE_SSH_PORT } else { "22" }
$RemoteBinDir = if ($env:REMOTE_BIN_DIR) { $env:REMOTE_BIN_DIR } else { "/home/anyadmin/app" }

# Standard SSH/SCP options - BatchMode=yes makes it non-interactive
$CommonSshArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-p", $RemotePort, "-i", $KeyFile)
$CommonScpArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-P", $RemotePort, "-i", $KeyFile)

Write-Host "Starting Agent Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 1. Compile Agent
Write-Host "[1/6] Compiling Agent for Linux/AMD64..." -ForegroundColor Yellow
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

# 2. Install Node.js (Required for Frontend)
Write-Host "[2/6] Ensuring Node.js is installed on Remote..." -ForegroundColor Yellow
$NodeCheckCmd = "node -v || (curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs)"
try {
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" $NodeCheckCmd
    Write-Host "Node.js is ready." -ForegroundColor Green
} catch {
    Write-Warning "Failed to ensure Node.js installation: $_"
}

# 3. Stop Remote Agent
Write-Host "[3/6] Stopping Remote Agent..." -ForegroundColor Yellow
try {
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "pkill -9 $AgentName || true"
    Write-Host "Remote agent stopped (if running)." -ForegroundColor Green
} catch {
    Write-Warning "Failed to stop agent or connection issue: $_"
}

# 4. Upload Agent
Write-Host "[4/6] Uploading Agent Binary..." -ForegroundColor Yellow
try {
    scp @CommonScpArgs "$BackendDir\dist\$AgentName" "$RemoteUser@$RemoteHost`:$RemoteBinDir/$AgentName"
    if ($LASTEXITCODE -ne 0) { throw "SCP failed" }
    Write-Host "Upload successful." -ForegroundColor Green
} catch {
    throw "Upload failed: $_"
}

# 5 Upload Docker Configurations
Write-Host "[5/6] Uploading Docker Configurations..." -ForegroundColor Yellow
$LocalDockerDir = "$BackendDir\deployments\dockers\yaml"
$RemoteDockerDir = "/home/anyadmin/docker"
try {
    # Ensure directory exists on remote
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "mkdir -p $RemoteDockerDir && chown anyadmin:anyadmin $RemoteDockerDir"
    
    # Upload files
    scp @CommonScpArgs "$LocalDockerDir\*" "$LocalDockerDir\.[!.]*" "$RemoteUser@$RemoteHost`:$RemoteDockerDir/" 
    
    # Set ownership for uploaded files
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "chown -R anyadmin:anyadmin $RemoteDockerDir"
    
    Write-Host "Docker configurations uploaded successfully." -ForegroundColor Green
} catch {
    Write-Warning "Failed to upload Docker configurations: $_"
}

# 6. Start Remote Agent
Write-Host "[6/6] Starting Remote Agent..." -ForegroundColor Yellow
$StartCmd = "chmod +x $RemoteBinDir/$AgentName && runuser -l anyadmin -c 'cd $RemoteBinDir && (nohup ./$AgentName -config config.json -log /home/anyadmin/logs/agent.log > /home/anyadmin/logs/agent.log 2>&1 < /dev/null &)'"
try {
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" $StartCmd
    if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
    Write-Host "Agent started successfully." -ForegroundColor Green
} catch {
    throw "Failed to start agent: $_"
}

Write-Host "Deployment Complete!" -ForegroundColor Cyan
