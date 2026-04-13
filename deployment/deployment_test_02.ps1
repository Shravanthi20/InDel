$ErrorActionPreference = 'Continue'
Add-Type -AssemblyName System.Net.Http

$baseCore = 'https://indel-backend.onrender.com'
$baseWorker = 'https://indel-worker-gateway.onrender.com'
$baseInsurer = 'https://indel-insurer-gateway.onrender.com'
$basePlatform = 'https://indel-bvvq.onrender.com'
$basePremiumML = 'https://indel-ml-premium.onrender.com'
$baseFraudML = 'https://indel-ml-fraud.onrender.com'
$baseForecastML = 'https://indel-ml-forecast.onrender.com'

$premiumPayload = @'
{
  "worker_id": "wkr_demo_001",
  "zone_id": "zone_tambaram_chennai",
  "city": "Chennai",
  "state": "Tamil Nadu",
  "zone_type": "urban",
  "vehicle_type": "two_wheeler",
  "season": "monsoon",
  "experience_days": 240,
  "avg_daily_orders": 18,
  "avg_daily_earnings": 1450,
  "active_hours_per_day": 9,
  "rainfall_mm": 22,
  "aqi": 126,
  "temperature": 31,
  "humidity": 79,
  "order_volatility": 0.34,
  "earnings_volatility": 0.29,
  "recent_disruption_rate": 0.21
}
'@

$premiumBatchPayload = @'
[
  {
    "worker_id": "wkr_demo_001",
    "zone_id": "zone_tambaram_chennai",
    "city": "Chennai",
    "state": "Tamil Nadu",
    "zone_type": "urban",
    "vehicle_type": "two_wheeler",
    "season": "monsoon",
    "experience_days": 240,
    "avg_daily_orders": 18,
    "avg_daily_earnings": 1450,
    "active_hours_per_day": 9,
    "rainfall_mm": 22,
    "aqi": 126,
    "temperature": 31,
    "humidity": 79,
    "order_volatility": 0.34,
    "earnings_volatility": 0.29,
    "recent_disruption_rate": 0.21
  },
  {
    "worker_id": "wkr_demo_002",
    "zone_id": "zone_rohini_delhi",
    "city": "Delhi",
    "state": "Delhi",
    "zone_type": "urban",
    "vehicle_type": "scooter",
    "season": "summer",
    "experience_days": 180,
    "avg_daily_orders": 14,
    "avg_daily_earnings": 1100,
    "active_hours_per_day": 8,
    "rainfall_mm": 2,
    "aqi": 88,
    "temperature": 35,
    "humidity": 44,
    "order_volatility": 0.21,
    "earnings_volatility": 0.18,
    "recent_disruption_rate": 0.07
  }
]
'@

$fraudPayload = @'
{
  "claim_id": 101,
  "worker_id": 10,
  "zone_id": 1,
  "claim_amount": 650,
  "baseline_earnings": 1200,
  "disruption_type": "heavy_rain",
  "gps_in_zone": true,
  "deliveries_during_disruption": 1,
  "zone_avg_claim_amount": 520,
  "worker_history": {
    "total_claims_last_8_weeks": 1,
    "avg_claim_amount": 480,
    "earnings_variance": 0.2,
    "zone_change_count": 0,
    "days_active": 65,
    "delivery_attempt_rate": 0.84
  }
}
'@

$forecastPayload = '{"zone_id":1}'

$tests = @(
    @{service='core'; category='health'; method='GET'; base=$baseCore; path='/health'; expected='200'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/policy/weekly-cycle/run'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/claims/generate-for-disruption/1'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/claims/auto-process/1'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/payouts/queue/1'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/payouts/process'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='GET'; base=$baseCore; path='/api/v1/internal/payouts/reconciliation'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/api/v1/internal/data/synthetic/generate'; body='{}'; expected='200/4xx/5xx'},
    @{service='core'; category='operations'; method='POST'; base=$baseCore; path='/internal/v1/claims/1/payout'; body='{}'; expected='200/4xx/5xx'},

    @{service='worker-gateway'; category='health'; method='GET'; base=$baseWorker; path='/health'; expected='200'},
    @{service='worker-gateway'; category='health'; method='GET'; base=$baseWorker; path='/api/v1/health'; expected='200'},
    @{service='worker-gateway'; category='health'; method='GET'; base=$baseWorker; path='/api/v1/status'; expected='200'},
    @{service='worker-gateway'; category='auth'; method='POST'; base=$baseWorker; path='/api/v1/auth/register'; body='{}'; expected='400'},
    @{service='worker-gateway'; category='auth'; method='POST'; base=$baseWorker; path='/api/v1/auth/login'; body='{}'; expected='400'},
    @{service='worker-gateway'; category='auth'; method='POST'; base=$baseWorker; path='/api/v1/auth/otp/send'; body='{}'; expected='400'},
    @{service='worker-gateway'; category='auth'; method='POST'; base=$baseWorker; path='/api/v1/auth/otp/verify'; body='{}'; expected='400'},
    @{service='worker-gateway'; category='worker'; method='POST'; base=$baseWorker; path='/api/v1/worker/onboard'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='worker'; method='GET'; base=$baseWorker; path='/api/v1/worker/profile'; expected='401/403'},
    @{service='worker-gateway'; category='worker'; method='PUT'; base=$baseWorker; path='/api/v1/worker/profile'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='policy'; method='GET'; base=$baseWorker; path='/api/v1/worker/policy'; expected='401/403'},
    @{service='worker-gateway'; category='policy'; method='POST'; base=$baseWorker; path='/api/v1/worker/policy/enroll'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='policy'; method='PUT'; base=$baseWorker; path='/api/v1/worker/policy/pause'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='policy'; method='PUT'; base=$baseWorker; path='/api/v1/worker/policy/cancel'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='policy'; method='GET'; base=$baseWorker; path='/api/v1/worker/policy/premium'; expected='401/403'},
    @{service='worker-gateway'; category='policy'; method='POST'; base=$baseWorker; path='/api/v1/worker/policy/premium/pay'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='earnings'; method='GET'; base=$baseWorker; path='/api/v1/worker/earnings'; expected='401/403'},
    @{service='worker-gateway'; category='earnings'; method='GET'; base=$baseWorker; path='/api/v1/worker/earnings/history'; expected='401/403'},
    @{service='worker-gateway'; category='earnings'; method='GET'; base=$baseWorker; path='/api/v1/worker/earnings/baseline'; expected='401/403'},
    @{service='worker-gateway'; category='claims'; method='GET'; base=$baseWorker; path='/api/v1/worker/claims'; expected='401/403'},
    @{service='worker-gateway'; category='claims'; method='GET'; base=$baseWorker; path='/api/v1/worker/claims/1'; expected='401/403/404'},
    @{service='worker-gateway'; category='wallet'; method='GET'; base=$baseWorker; path='/api/v1/worker/wallet'; expected='401/403'},
    @{service='worker-gateway'; category='payouts'; method='GET'; base=$baseWorker; path='/api/v1/worker/payouts'; expected='401/403'},
    @{service='worker-gateway'; category='orders'; method='GET'; base=$baseWorker; path='/api/v1/worker/orders'; expected='401/403'},
    @{service='worker-gateway'; category='orders'; method='GET'; base=$baseWorker; path='/api/v1/worker/orders/available'; expected='200/401/403'},
    @{service='worker-gateway'; category='orders'; method='GET'; base=$baseWorker; path='/api/v1/worker/orders/assigned'; expected='401/403'},
    @{service='worker-gateway'; category='orders'; method='GET'; base=$baseWorker; path='/api/v1/worker/orders/1'; expected='401/403/404'},
    @{service='worker-gateway'; category='orders'; method='PUT'; base=$baseWorker; path='/api/v1/worker/orders/1/accept'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='orders'; method='PUT'; base=$baseWorker; path='/api/v1/worker/orders/1/picked-up'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='orders'; method='PUT'; base=$baseWorker; path='/api/v1/worker/orders/1/delivered'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='orders'; method='POST'; base=$baseWorker; path='/api/v1/worker/orders/1/code/send'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='verification'; method='POST'; base=$baseWorker; path='/api/v1/worker/fetch-verification/send-code'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='verification'; method='POST'; base=$baseWorker; path='/api/v1/worker/fetch-verification/verify'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='config'; method='GET'; base=$baseWorker; path='/api/v1/worker/zone-config'; expected='401/403'},
    @{service='worker-gateway'; category='session'; method='GET'; base=$baseWorker; path='/api/v1/worker/session/1'; expected='401/403/404'},
    @{service='worker-gateway'; category='session'; method='GET'; base=$baseWorker; path='/api/v1/worker/session/1/deliveries'; expected='401/403/404'},
    @{service='worker-gateway'; category='session'; method='GET'; base=$baseWorker; path='/api/v1/worker/session/1/fraud-signals'; expected='401/403/404'},
    @{service='worker-gateway'; category='session'; method='PUT'; base=$baseWorker; path='/api/v1/worker/session/1/end'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='notifications'; method='GET'; base=$baseWorker; path='/api/v1/worker/notifications'; expected='401/403'},
    @{service='worker-gateway'; category='notifications'; method='PUT'; base=$baseWorker; path='/api/v1/worker/notifications/preferences'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='notifications'; method='POST'; base=$baseWorker; path='/api/v1/worker/notifications/fcm-token'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/trigger-disruption'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/settle-earnings'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/reset-zone'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/reset'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/assign-orders'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/simulate-orders'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/simulate-deliveries'; body='{}'; expected='401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/orders/publisher/initiate'; body='{}'; expected='200/401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/orders/publisher/ack'; body='{}'; expected='200/401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='GET'; base=$baseWorker; path='/api/v1/demo/orders/publisher/status'; expected='200/401/403'},
    @{service='worker-gateway'; category='demo'; method='POST'; base=$baseWorker; path='/api/v1/demo/orders/ingest'; body='{}'; expected='200/401/403/4xx'},
    @{service='worker-gateway'; category='demo'; method='GET'; base=$baseWorker; path='/api/v1/demo/orders/search'; expected='200/401/403'},
    @{service='worker-gateway'; category='demo'; method='GET'; base=$baseWorker; path='/api/v1/demo/orders/available'; expected='200/401/403'},
    @{service='worker-gateway'; category='demo'; method='GET'; base=$baseWorker; path='/api/v1/demo/deliveries'; expected='200/401/403'},

    @{service='insurer-gateway'; category='health'; method='GET'; base=$baseInsurer; path='/health'; expected='200'},
    @{service='insurer-gateway'; category='overview'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/overview'; expected='200/4xx'},
    @{service='insurer-gateway'; category='overview'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/loss-ratio'; expected='200/4xx'},
    @{service='insurer-gateway'; category='claims'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/claims'; expected='200/4xx'},
    @{service='insurer-gateway'; category='claims'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/claims/fraud-queue'; expected='200/4xx'},
    @{service='insurer-gateway'; category='claims'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/claims/1'; expected='200/4xx/404'},
    @{service='insurer-gateway'; category='claims'; method='POST'; base=$baseInsurer; path='/api/v1/insurer/claims/1/review'; body='{}'; expected='200/4xx/401/403'},
    @{service='insurer-gateway'; category='forecast'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/forecast'; expected='200/4xx'},
    @{service='insurer-gateway'; category='pool'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/pool/health'; expected='200/4xx'},
    @{service='insurer-gateway'; category='maintenance'; method='GET'; base=$baseInsurer; path='/api/v1/insurer/maintenance-checks'; expected='200/4xx/401/403'},
    @{service='insurer-gateway'; category='maintenance'; method='POST'; base=$baseInsurer; path='/api/v1/insurer/maintenance-checks/1/respond'; body='{}'; expected='200/4xx/401/403'},

    @{service='platform-gateway'; category='health'; method='GET'; base=$basePlatform; path='/health'; expected='200'},
    @{service='platform-gateway'; category='platform'; method='GET'; base=$basePlatform; path='/api/v1/platform/workers'; expected='200/4xx'},
    @{service='platform-gateway'; category='platform'; method='GET'; base=$basePlatform; path='/api/v1/platform/zones'; expected='200/4xx'},
    @{service='platform-gateway'; category='platform'; method='POST'; base=$basePlatform; path='/api/v1/platform/webhooks/order/assigned'; body='{}'; expected='200/4xx/5xx'},
    @{service='platform-gateway'; category='platform'; method='POST'; base=$basePlatform; path='/api/v1/platform/webhooks/order/completed'; body='{}'; expected='200/4xx/5xx'},
    @{service='platform-gateway'; category='platform'; method='POST'; base=$basePlatform; path='/api/v1/platform/webhooks/order/cancelled'; body='{}'; expected='200/4xx/5xx'},
    @{service='platform-gateway'; category='platform'; method='POST'; base=$basePlatform; path='/api/v1/platform/webhooks/external-signal'; body='{}'; expected='200/4xx/5xx'},
    @{service='platform-gateway'; category='platform'; method='GET'; base=$basePlatform; path='/api/v1/platform/zones/health'; expected='200/4xx'},
    @{service='platform-gateway'; category='platform'; method='GET'; base=$basePlatform; path='/api/v1/platform/disruptions'; expected='200/4xx'},
    @{service='platform-gateway'; category='demo'; method='POST'; base=$basePlatform; path='/api/v1/demo/trigger-disruption'; body='{}'; expected='200/4xx/5xx'},

    @{service='premium-ml'; category='health'; method='GET'; base=$basePremiumML; path='/health'; expected='200'},
    @{service='premium-ml'; category='premium'; method='POST'; base=$basePremiumML; path='/ml/v1/premium/calculate'; body=$premiumPayload; expected='200'},
    @{service='premium-ml'; category='premium'; method='POST'; base=$basePremiumML; path='/ml/v1/premium/batch-calculate'; body=$premiumBatchPayload; expected='200'},

    @{service='fraud-ml'; category='health'; method='GET'; base=$baseFraudML; path='/health'; expected='200'},
    @{service='fraud-ml'; category='fraud'; method='POST'; base=$baseFraudML; path='/ml/v1/fraud/score'; body=$fraudPayload; expected='200'},

    @{service='forecast-ml'; category='health'; method='GET'; base=$baseForecastML; path='/health'; expected='200'},
    @{service='forecast-ml'; category='forecast'; method='POST'; base=$baseForecastML; path='/forecast'; body=$forecastPayload; expected='200'}
)

$client = [System.Net.Http.HttpClient]::new()
$client.Timeout = [TimeSpan]::FromSeconds(60)
$results = @()

foreach ($t in $tests) {
    $url = $t.base + $t.path
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::new($t.method), $url)
        if ($t.ContainsKey('body')) {
            $request.Content = [System.Net.Http.StringContent]::new($t.body, [System.Text.Encoding]::UTF8, 'application/json')
        }
        $resp = $client.SendAsync($request).GetAwaiter().GetResult()
        $sw.Stop()
        $content = $resp.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        if ($content.Length -gt 240) { $content = $content.Substring(0, 240) }
        $results += [PSCustomObject]@{
            service = $t.service
            category = $t.category
            method = $t.method
            path = $t.path
            expected = $t.expected
            status = [int]$resp.StatusCode.value__
            ms = [int]$sw.ElapsedMilliseconds
            sample = $content
        }
        $request.Dispose()
        $resp.Dispose()
    }
    catch {
        $sw.Stop()
        $msg = $_.Exception.Message
        if ($msg.Length -gt 240) { $msg = $msg.Substring(0, 240) }
        $results += [PSCustomObject]@{
            service = $t.service
            category = $t.category
            method = $t.method
            path = $t.path
            expected = $t.expected
            status = $null
            ms = [int]$sw.ElapsedMilliseconds
            sample = $msg
        }
    }
}

$client.Dispose()

$runTime = Get-Date -Format 'yyyy-MM-dd HH:mm:ss K'
$total = $results.Count
$ok200 = @($results | Where-Object { $_.status -eq 200 }).Count
$status4xx = @($results | Where-Object { $_.status -ge 400 -and $_.status -lt 500 }).Count
$status5xx = @($results | Where-Object { $_.status -ge 500 -and $_.status -lt 600 }).Count

Write-Output "RUN_TIME=$runTime"
Write-Output "TOTAL=$total"
Write-Output "OK_200=$ok200"
Write-Output "STATUS_4XX=$status4xx"
Write-Output "STATUS_5XX=$status5xx"

$criticalHealthFailures = @(
  $results |
  Where-Object { $_.category -eq 'health' -and $_.expected -eq '200' -and $_.status -ne 200 }
)

if ($criticalHealthFailures.Count -gt 0) {
  Write-Output "CRITICAL_HEALTH_FAIL_COUNT=$($criticalHealthFailures.Count)"
  $criticalHealthFailures | ForEach-Object {
    Write-Output ("CRITICAL_HEALTH_FAIL={0} {1} -> status={2} sample={3}" -f $_.service, $_.path, $_.status, ($_.sample -replace '\r|\n', ' '))
  }
}

Write-Output "DETAILS_BEGIN"
$results | ForEach-Object {
    $sample = ($_.sample -replace '\|','/' -replace '\r|\n',' ')
    Write-Output ("{0}|{1}|{2}|{3}|{4}|{5}|{6}|{7}" -f $_.service,$_.category,$_.method,$_.path,$_.status,$_.ms,$_.expected,$sample)
}
Write-Output "DETAILS_END"

if ($criticalHealthFailures.Count -gt 0) {
  exit 1
}
