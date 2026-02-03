# Helper function to stop process by port
function Stop-ProcessByPort {
    param (
        [int]$Port
    )

    $ErrorActionPreference = "SilentlyContinue"
    # Get PIDs listening on the specific port
    $connections = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
    
    if ($connections) {
        $pidsToStop = $connections | Select-Object -ExpandProperty OwningProcess -Unique
        foreach ($id in $pidsToStop) {
            # Skip system idle process (PID 0)
            if ($id -eq 0) { continue }

            try {
                $proc = Get-Process -Id $id -ErrorAction Stop
                Stop-Process -Id $id -Force
                Write-Host "Stopped process on port $Port (PID: $id, Name: $($proc.ProcessName))" -ForegroundColor Green
            } catch {
                Write-Host "Failed to stop process on port $Port (PID: $id) - It may not exist or access is denied" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "No process found listening on port $Port" -ForegroundColor Yellow
    }
}

# Stop ports 3000 and 8080
Stop-ProcessByPort 3000
Stop-ProcessByPort 8080