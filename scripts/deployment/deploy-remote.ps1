# scripts/deployment/deploy-remote.ps1
# This script deploys the Node.js frontend and Go backend to a remote Ubuntu 22.04 server.
# It ensures Node.js and Go are installed on the remote host before building.
# Execution is performed via ROOT account, with tasks delegated to 'anyadmin' user.

$ErrorActionPreference = "Stop"

# --- HELPER FUNCTIONS ---
function Load-Env {
    param ([string]$Path = ".env")
    if (Test-Path $Path) {
        Write-Host "Loading local environment from $Path" -ForegroundColor Gray
        Get-Content $Path | ForEach-Object {
            if ($_ -match '^([^#=]+)=(.*)$') {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                # Remove quotes if present
                $value = $value -replace '^["'']|["'']$' , ''
                [System.Environment]::SetEnvironmentVariable($name, $value, [System.EnvironmentVariableTarget]::Process)
            }
        }
    }
}

# Load local .env if it exists to populate REMOTE_USER/REMOTE_HOST
Load-Env -Path ".env"

# --- CONFIGURATION ---
$remoteUser = if ($env:REMOTE_USER) { $env:REMOTE_USER } else { "root" }
$remoteHost = if ($env:REMOTE_HOST) { $env:REMOTE_HOST } else { "172.25.208.100" }

$keyFile = "backend/keys/id_rsa"
$goVersion = "1.25.6"
# ---------------------

# SSH/SCP Common Arguments
$sshOpts = @("-i", $keyFile, "-o", "StrictHostKeyChecking=no", "-o", "BatchMode=yes")
$scpOpts = @("-i", $keyFile, "-o", "StrictHostKeyChecking=no", "-o", "BatchMode=yes")

function Write-ProgressMsg([string]$msg) {
    Write-Host "`n>>> $msg" -ForegroundColor Cyan
}

function Execute-Remote([string]$command) {
    ssh @sshOpts "$remoteUser@$remoteHost" "$command"
    if ($LASTEXITCODE -ne 0) {
        throw "Remote command failed with exit code ${LASTEXITCODE}: $command"
    }
}

function Execute-AsAnyadmin([string]$command, [bool]$loadEnv = $false) {
    $finalCommand = $command
    if ($loadEnv) {
        # Source .env and export all variables (set -a) before running the command on the remote host.
        # We use single quotes or escape the command to prevent local PowerShell evaluation.
        $finalCommand = "set -a; [ -f ~/src/.env ] && . ~/src/.env; set +a; $command"
    }
    # Escape single quotes in the command for use inside the runuser -c '...' string
    $escapedCommand = $finalCommand.Replace("'", "'\''")
    Execute-Remote "runuser -l anyadmin -c '$escapedCommand'"
}

try {
    Write-ProgressMsg "Starting deployment to $remoteUser@$remoteHost..."

    # 1. Ensure anyadmin user exists
    Write-ProgressMsg "[1/9] Ensuring 'anyadmin' user exists..."
    $userCmd = "id -u anyadmin >/dev/null 2>&1 || (useradd -m -s /bin/bash anyadmin && echo 'anyadmin created') && " +
               "usermod -aG sudo anyadmin 2>/dev/null || true"
    Execute-Remote $userCmd

    # 2. Prepare remote directories
    Write-ProgressMsg "[2/9] Preparing remote directories..."
    Execute-Remote "mkdir -p /home/anyadmin/src/frontend /home/anyadmin/src/backend /home/anyadmin/logs"
    Execute-Remote "chown -R anyadmin:anyadmin /home/anyadmin"

    # 3. Copy .env file
    Write-ProgressMsg "[3/9] Copying .env file..."
    if (Test-Path ".env") {
        scp @scpOpts ".env" "$remoteUser@$remoteHost`:/home/anyadmin/src/.env"
        # Fix Windows line endings (CRLF -> LF) to prevent shell sourcing errors
        Execute-Remote "sed -i 's/\r$//' /home/anyadmin/src/.env"
        # Distribute and set ownership
        Execute-Remote "cp /home/anyadmin/src/.env /home/anyadmin/src/frontend/.env"
        Execute-Remote "cp /home/anyadmin/src/.env /home/anyadmin/src/backend/.env"
        Execute-Remote "chown anyadmin:anyadmin /home/anyadmin/src/.env /home/anyadmin/src/frontend/.env /home/anyadmin/src/backend/.env"
    } else {
        Write-Host "Warning: .env file not found locally." -ForegroundColor Yellow
    }

    # 4. Copy source code
    Write-ProgressMsg "[4/9] Packaging and copying frontend (skipping node_modules)..."
    # Copy root package.json as well because it might contain build dependencies
    scp @scpOpts "package.json" "$remoteUser@$remoteHost`:/home/anyadmin/src/package.json"
    
    $feArchive = "frontend_src.tar.gz"
    tar -czf $feArchive --exclude="node_modules" -C frontend .
    scp @scpOpts $feArchive "$remoteUser@$remoteHost`:/tmp/$feArchive"
    Execute-Remote "tar -xzf /tmp/$feArchive -C /home/anyadmin/src/frontend && rm /tmp/$feArchive"
    Remove-Item $feArchive

    Write-ProgressMsg "[4/9] Packaging and copying backend (skipping deployments/tars)..."
    $beArchive = "backend_src.tar.gz"
    tar -czf $beArchive --exclude="deployments/tars" -C backend .
    scp @scpOpts $beArchive "$remoteUser@$remoteHost`:/tmp/$beArchive"
    Execute-Remote "tar -xzf /tmp/$beArchive -C /home/anyadmin/src/backend && rm /tmp/$beArchive"
    Remove-Item $beArchive

    Execute-Remote "chown -R anyadmin:anyadmin /home/anyadmin/src"

    # 5. Setup Remote Environment (Node.js & Go)
    Write-ProgressMsg "[5/9] Setting up remote environment (as root)..."
    $nodeSetup = "if ! command -v node >/dev/null 2>&1 || ! node -v | grep -q 'v22'; then " +
                 "echo 'Node.js 22 not found. Installing...'; " +
                 "apt-get update && apt-get install -y curl gnupg && " +
                 "curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && " +
                 "apt-get install -y nodejs; " +
                 "else " +
                 "echo 'Node.js is already installed ('$(node -v)').'; " +
                 "fi"
    Execute-Remote $nodeSetup

    # Check if Go is available
    ssh @sshOpts "$remoteUser@$remoteHost" "command -v go >/dev/null 2>&1"
    if ($LASTEXITCODE -ne 0) {
        $localGoTar = "backend/deployments/tars/os/ubuntu/amd64/jammy/go$goVersion.linux-amd64.tar.gz"
        if (Test-Path $localGoTar) {
            Write-ProgressMsg "Installing Go $goVersion from local tarball..."
            scp @scpOpts $localGoTar "$remoteUser@$remoteHost`:/tmp/go.tar.gz"
            $goInstall = "mkdir -p /home/anyadmin/tmp/go-install && " +
                         "tar -C /home/anyadmin/tmp/go-install -xzf /tmp/go.tar.gz && " +
                         "ln -sf /home/anyadmin/tmp/go-install/go/bin/go /usr/bin/go && " +
                         "ln -sf /home/anyadmin/tmp/go-install/go/bin/gofmt /usr/bin/gofmt && " +
                         "rm /tmp/go.tar.gz"
            Execute-Remote $goInstall
        } else {
            Write-ProgressMsg "Installing Go via apt-get..."
            Execute-Remote "apt-get update && apt-get install -y golang-go"
        }
    } else {
        Write-Host "Go is already installed." -ForegroundColor Gray
    }

    # 6. Build Frontend
    Write-ProgressMsg "[6/9] Building Frontend as 'anyadmin'..."
    Execute-AsAnyadmin "cd ~/src/frontend && npm install && npm run build" $true

    # 7. Build Backend
    Write-ProgressMsg "[7/9] Building Backend as 'anyadmin'..."
    Execute-AsAnyadmin "cd ~/src/backend && go build -o anyadmin-server ./cmd/server/main.go" $true

    # 8. Finalize
    Write-ProgressMsg "[8/9] Finalizing permissions..."
    Execute-Remote "chown -R anyadmin:anyadmin /home/anyadmin"

    # 9. Create and run start script
    Write-ProgressMsg "[9/9] Creating and running start script..."
    $startScript = @"
#!/bin/bash
# Stop existing processes
pkill -f "anyadmin-server" || true
pkill -f "node app.js" || true

# Wait a moment for ports to be released
sleep 2

# Source .env if exists
set -a; [ -f /home/anyadmin/src/.env ] && . /home/anyadmin/src/.env; set +a

# Start Backend
cd /home/anyadmin/src/backend
nohup ./anyadmin-server > /home/anyadmin/logs/backend.log 2>&1 &
echo "Backend started with PID $!"

# Start Frontend
cd /home/anyadmin/src/frontend
nohup node app.js > /home/anyadmin/logs/frontend.log 2>&1 &
echo "Frontend started with PID $!"
"@
    
    # Save script to a temporary local file
    $tmpScript = "start-anyadmin.sh"
    # Use ASCII to avoid BOM issues in PowerShell 5.1
    $startScript | Out-File -FilePath $tmpScript -Encoding Ascii
    
    # Copy to remote and set permissions
    scp @scpOpts $tmpScript "$remoteUser@$remoteHost`:/home/anyadmin/start-anyadmin.sh"
    # Fix Windows line endings (CRLF -> LF) and remove potential BOM (just in case)
    Execute-Remote "sed -i '1s/^\xEF\xBB\xBF//; s/\r$//' /home/anyadmin/start-anyadmin.sh"
    Execute-Remote "chmod +x /home/anyadmin/start-anyadmin.sh && chown anyadmin:anyadmin /home/anyadmin/start-anyadmin.sh"
    
    # Execute as anyadmin
    Execute-AsAnyadmin "/home/anyadmin/start-anyadmin.sh"
    
    Remove-Item $tmpScript

    Write-ProgressMsg "Deployment completed successfully!"
}
catch {
    Write-Host "`nDeployment FAILED: $($_.Exception.Message)" -ForegroundColor Red
    if (Test-Path "frontend_src.tar.gz") { Remove-Item "frontend_src.tar.gz" }
    if (Test-Path "backend_src.tar.gz") { Remove-Item "backend_src.tar.gz" }
    if (Test-Path "start-anyadmin.sh") { Remove-Item "start-anyadmin.sh" }
    exit 1
}
