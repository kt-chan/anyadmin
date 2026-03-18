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
$RemoteBinDir = if ($env:REMOTE_BIN_DIR) { $env:REMOTE_BIN_DIR } else { "/home/anyadmin/bin" }
$RemoteSrcDir = if ($env:REMOTE_SRC_DIR) { $env:REMOTE_SRC_DIR } else { "/home/anyadmin/src" }

# Standard SSH/SCP options - BatchMode=yes makes it non-interactive
$CommonSshArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-p", $RemotePort, "-i", $KeyFile)
$CommonScpArgs = @("-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-P", $RemotePort, "-i", $KeyFile)

Write-Host "Starting Agent Deployment to $RemoteHost (SSH Port: $RemotePort)..." -ForegroundColor Cyan

# 1. Sync Project Source to Remote
Write-Host "[1/6] Syncing Project Source to Remote..." -ForegroundColor Yellow
Sync-RemoteSource `
    -LocalPath $ProjectRoot `
    -RemotePath $RemoteSrcDir `
    -ArchiveName "project_src" `
    -RemoteUser $RemoteUser `
    -RemoteHost $RemoteHost `
    -CommonSshArgs $CommonSshArgs `
    -CommonScpArgs $CommonScpArgs

# 2. Build Agent on Remote Host
Write-Host "[2/6] Building Agent on Remote Host..." -ForegroundColor Yellow
# Ensure binary directory exists, source Go profile, and build
$BuildCmd = "source /etc/profile.d/go.sh 2>/dev/null || true; mkdir -p $RemoteBinDir && cd $RemoteSrcDir/backend && go build -o $RemoteBinDir/$AgentName ./cmd/agent/main.go"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$BuildCmd"
if ($LASTEXITCODE -ne 0) { throw "Remote compilation failed" }
Write-Host "Remote compilation successful." -ForegroundColor Green

# 3. Stop Remote Agent
Write-Host "[3/6] Stopping Remote Agent..." -ForegroundColor Yellow
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "pkill -9 $AgentName || true"
Write-Host "Remote agent stopped (if running)." -ForegroundColor Green

# 4. Upload Docker Configurations
Write-Host "[4/6] Uploading Docker Configurations and Config..." -ForegroundColor Yellow
$LocalDockerDir = "$BackendDir\deployments\dockers\yaml"
$RemoteDockerDir = "/home/anyadmin/docker"
# Ensure directory exists and upload
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "mkdir -p $RemoteDockerDir $RemoteBinDir && chown anyadmin:anyadmin $RemoteDockerDir $RemoteBinDir"
scp @CommonScpArgs "$LocalDockerDir\*" "$LocalDockerDir\.[!.]*" "$RemoteUser@$RemoteHost`:$RemoteDockerDir/" 

# Generate config.json for the agent pointing BACK to the server
# In this deployment, the agent's mgmt_host should point to the MgmtHost from .env
$MgmtHost = if ($env:MgmtHost) { $env:MgmtHost } else { $RemoteHost }
$MgmtPort = if ($env:MgmtPort) { $env:MgmtPort } else { "8080" }

$AgentConfig = @{
    mgmt_host       = $MgmtHost
    mgmt_port       = $MgmtPort
    node_ip         = $RemoteHost
    node_port       = "8082"
    deployment_time = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"
    log_file        = "/home/anyadmin/logs/agent.log"
} | ConvertTo-Json

$TempConfigPath = Join-Path $env:TEMP "agent_config.json"
[System.IO.File]::WriteAllText($TempConfigPath, $AgentConfig)
scp @CommonScpArgs $TempConfigPath "$RemoteUser@$RemoteHost`:$RemoteBinDir/config.json"
Remove-Item $TempConfigPath

ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "chown -R anyadmin:anyadmin $RemoteDockerDir $RemoteBinDir"
Write-Host "Configurations uploaded successfully." -ForegroundColor Green

# 5. Start Remote Agent
Write-Host "[5/6] Starting Remote Agent..." -ForegroundColor Yellow
$StartCmd = "ls -l $RemoteBinDir/$AgentName && chmod +x $RemoteBinDir/$AgentName && runuser -l anyadmin -c 'cd $RemoteBinDir && (nohup ./$AgentName -config config.json -log /home/anyadmin/logs/agent.log > /home/anyadmin/logs/agent.log 2>&1 < /dev/null &)'"
ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$StartCmd"
if ($LASTEXITCODE -ne 0) { throw "Start command failed" }
Write-Host "Agent started successfully." -ForegroundColor Green

Write-Host "Deployment Complete!" -ForegroundColor Cyan
