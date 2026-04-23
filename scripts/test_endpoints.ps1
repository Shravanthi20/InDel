# PowerShell script to test InDel API endpoints

$baseUrl = "https://indel-insurer-gateway.onrender.com/api/v1"

$endpoints = @(
    "/insurer/ledger?limit=10",
    "/insurer/money-exchange?level=A&zone=",
    "/insurer/users",
    "/platform/zones"
)

foreach ($endpoint in $endpoints) {
    $url = $baseUrl + $endpoint
    Write-Host "Testing: $url"
    try {
        $response = Invoke-WebRequest -Uri $url -Method GET -ErrorAction Stop
        Write-Host "Status: $($response.StatusCode)"
        Write-Host "Response: $($response.Content.Substring(0, [Math]::Min(500, $response.Content.Length)))..."
    } catch {
        Write-Host "Error: $($_.Exception.Message)"
        if ($_.Exception.Response -ne $null) {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $body = $reader.ReadToEnd()
            Write-Host "Body: $body"
        }
    }
    Write-Host "-----------------------------"
}
