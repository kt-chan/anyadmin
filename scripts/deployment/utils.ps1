# Load .env variables into environment
function Load-Env {
    param (
        [string]$Path = ".env"
    )
    if (Test-Path $Path) {
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

function Ensure-RemoteNodeJS {
    param (
        [string]$RemoteUser,
        [string]$RemoteHost,
        [array]$CommonSshArgs
    )
    Write-Host "Ensuring Node.js 20+ is installed on Remote..." -ForegroundColor Yellow
    # Check node version, if < 18 or not found, install 20
    $CheckCmd = "node -v | grep -E 'v(18|20|22|24)' || (curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs)"
    ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "$CheckCmd"
}

function Sync-RemoteSource {
    param (
        [string]$LocalPath,
        [string]$RemotePath,
        [string]$ArchiveName,
        [string]$RemoteUser,
        [string]$RemoteHost,
        [array]$CommonSshArgs,
        [array]$CommonScpArgs,
        [array]$Excludes = @("dist", "bin", "node_modules", ".git", "logs", "*.tar","*.tar.gz")
    )

    $TempArchive = Join-Path $env:TEMP "$ArchiveName.tar.gz"
    
    # 1. Create Tarball
    Write-Host "Creating local archive $TempArchive from $LocalPath..." -ForegroundColor Gray
    Push-Location $LocalPath
    try {
        if (Test-Path $TempArchive) { Remove-Item $TempArchive }
        $TarArgs = @("-czf", $TempArchive)
        foreach ($ex in $Excludes) {
            $TarArgs += "--exclude=$ex"
        }
        $TarArgs += "."
        tar @TarArgs
        if ($LASTEXITCODE -ne 0) { throw "Local tar failed" }
    } finally {
        Pop-Location
    }

    # 2. Calculate Local Hash
    $LocalHash = (Get-FileHash $TempArchive -Algorithm SHA256).Hash.ToLower()
    Write-Host "Local Hash: $LocalHash" -ForegroundColor Gray

    # 3. Get Remote Hash
    $RemoteArchiveFile = "$RemotePath/../$ArchiveName.tar.gz"
    $RemoteHash = ""
    try {
        $RemoteHashOutput = ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "sha256sum $RemoteArchiveFile 2>/dev/null | cut -d' ' -f1"
        if ($LASTEXITCODE -eq 0) {
            $RemoteHash = $RemoteHashOutput.Trim().ToLower()
        }
    } catch { }

    # 4. Compare and Sync
    if ($LocalHash -eq $RemoteHash) {
        Write-Host "Hashes match. Skipping upload for $ArchiveName." -ForegroundColor Green
    } else {
        Write-Host "Hashes differ or remote missing. Uploading $ArchiveName..." -ForegroundColor Yellow
        ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "mkdir -p $RemotePath"
        scp @CommonScpArgs $TempArchive "$RemoteUser@$RemoteHost`:$RemoteArchiveFile"
        ssh @CommonSshArgs "$RemoteUser@$RemoteHost" "cd $RemotePath && tar -xzf ../$ArchiveName.tar.gz"
        Write-Host "Sync successful for $ArchiveName." -ForegroundColor Green
    }
}
