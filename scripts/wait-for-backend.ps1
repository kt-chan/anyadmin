param (
    [string]$HostIP = "127.0.0.1",
    [int]$Port = 8080,
    [int]$MaxRetries = 30
)

Write-Host "Waiting for Backend on $HostIP`:$Port..." -ForegroundColor Cyan

$retryCount = 0
while ($retryCount -lt $MaxRetries) {
    try {
        $client = New-Object System.Net.Sockets.TcpClient($HostIP, $Port)
        if ($client.Connected) {
            $client.Close()
            Write-Host "Backend is ready!" -ForegroundColor Green
            exit 0
        }
    } catch {
        # Port not open yet
    }
    
    $retryCount++
    Start-Sleep -Seconds 1
}

Write-Error "Backend failed to start after $MaxRetries seconds."
exit 1
