param(
  [string]$Token
)

if ([string]::IsNullOrWhiteSpace($Token)) {
  throw "Token is required"
}

function Get-CodeFromId {
  param(
    [Parameter(Mandatory = $true)][string]$Id,
    [Parameter(Mandatory = $true)][int]$Seed,
    [Parameter(Mandatory = $true)][int]$Multiplier
  )

  $value = $Seed
  foreach ($ch in $Id.Trim().ToUpper().ToCharArray()) {
    $value = (($value * $Multiplier) + [int][char]$ch) % 9000
  }

  return ('{0:D4}' -f (1000 + $value))
}

$headers = @{ Authorization = "Bearer $Token" }

$avail = Invoke-RestMethod -Method GET -Uri 'http://localhost:8004/api/v1/worker/batches?limit=100' -Headers $headers -TimeoutSec 30
$batches = @($avail.batches)
if ($batches.Count -eq 0) {
  Write-Output 'SMOKE no batches'
  exit 0
}

$batch = $batches[0]
$orderIds = @($batch.orders | ForEach-Object { $_.orderId })
$pickupCode = Get-CodeFromId -Id ([string]$batch.batchId) -Seed 0 -Multiplier 31
$acceptBody = @{ orderIds = $orderIds; pickupCode = $pickupCode } | ConvertTo-Json -Depth 6
$accept = Invoke-RestMethod -Method PUT -Uri ("http://localhost:8004/api/v1/worker/batches/{0}/accept" -f $batch.batchId) -Headers $headers -ContentType 'application/json' -Body $acceptBody -TimeoutSec 30

if ([string]$batch.zoneLevel -eq 'A') {
  $firstOrderId = [string](@($batch.orders | Select-Object -First 1).orderId)
  if ([string]::IsNullOrWhiteSpace($firstOrderId)) {
    Write-Output ("SMOKE zone=A batch=" + $batch.batchId + " accept=" + $accept.message + " delivery=missing_order_id")
    exit 0
  }

  $zoneADeliveryCode = Get-CodeFromId -Id $firstOrderId -Seed 11 -Multiplier 41
  $delBody = @{ deliveryCode = $zoneADeliveryCode } | ConvertTo-Json
  $deliver = Invoke-RestMethod -Method PUT -Uri ("http://localhost:8004/api/v1/worker/batches/{0}/deliver" -f $batch.batchId) -Headers $headers -ContentType 'application/json' -Body $delBody -TimeoutSec 30
  Write-Output ("SMOKE zone=A batch=" + $batch.batchId + " accept=" + $accept.message + " deliver=" + $deliver.message + " batchCompleted=" + $deliver.batchCompleted + " remaining=" + $deliver.remainingOrders)
}
else {
  $deliveryCode = Get-CodeFromId -Id ([string]$batch.batchId) -Seed 7 -Multiplier 37
  $delBody = @{ deliveryCode = $deliveryCode } | ConvertTo-Json
  $deliver = Invoke-RestMethod -Method PUT -Uri ("http://localhost:8004/api/v1/worker/batches/{0}/deliver" -f $batch.batchId) -Headers $headers -ContentType 'application/json' -Body $delBody -TimeoutSec 30
  Write-Output ("SMOKE zone=" + $batch.zoneLevel + " batch=" + $batch.batchId + " accept=" + $accept.message + " deliver=" + $deliver.message + " batchCompleted=" + $deliver.batchCompleted)
}
