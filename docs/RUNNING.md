# Running InDel Locally

This guide covers the updated stack with:

- backend gateways (`core`, `worker-gateway`, `insurer-gateway`, `platform-gateway`)
- ML services (`premium-ml`, `fraud-ml`, `forecast-ml`)
- unified API gateway (`api-gateway` on `:8004`)
- Platform Dashboard (`platform-dashboard` on `:5175`)

## Prerequisites

- Docker Desktop (with Compose v2)
- Ports available: `5175`, `8000`, `8001`, `8002`, `8003`, `8004`, `9001`, `9002`, `9003`

## 1. Start Demo Stack (Recommended)

From repository root:

```powershell
Copy-Item .env.demo.example .env -Force
docker compose -f docker-compose.demo.yml down -v
docker compose -f docker-compose.demo.yml up -d --build
```

The `.env.demo.example` values are the known-good baseline for demo-day stability.

Security-related environment notes:

- Set `INDEL_ALLOWED_ORIGINS` to the UI origins that are allowed to call backend APIs.
- Set `INDEL_DEMO_RESET_KEY` and send it as `X-Demo-Reset-Key` when calling `POST /api/v1/demo/reset` without bearer auth.
- Set `INDEL_DEMO_ALLOWED_ROLES` for demo operations (trigger/simulate/settle).
- Set `INDEL_DEMO_DESTRUCTIVE_ROLES` for destructive demo operations (`reset-zone`, destructive `reset`).
- Set `INDEL_CORE_INTERNAL_ALLOWED_ROLES` for `/api/v1/internal/*` and `/internal/v1/*` core operator endpoints.
- Set `INDEL_PLATFORM_OPERATOR_ALLOWED_ROLES` for `/api/v1/platform/demo/*` control endpoints.
- Set `INDEL_PLATFORM_WEBHOOK_ALLOWED_ROLES` or `INDEL_PLATFORM_WEBHOOK_KEY` for `/api/v1/platform/webhooks/*` endpoints.

## 2. Start Main Stack (Alternative)

From repository root:

```powershell
docker compose -f docker-compose.yml down -v
docker compose -f docker-compose.yml up -d --build
```

## 3. Verify Services

```powershell
$urls = @(
  'http://localhost:8000/health',
  'http://localhost:8001/api/v1/health',
  'http://localhost:8001/api/v1/status',
  'http://localhost:8001/health',
  'http://localhost:8002/health',
  'http://localhost:8002/api/v1/health',
  'http://localhost:8002/api/v1/status',
  'http://localhost:8003/health',
  'http://localhost:8003/api/v1/health',
  'http://localhost:8003/api/v1/status',
  'http://localhost:8004/health',
  'http://localhost:9001/health',
  'http://localhost:9002/health',
  'http://localhost:9003/health'
)
foreach ($u in $urls) {
  try {
    $r = Invoke-WebRequest -Uri $u -UseBasicParsing -TimeoutSec 8
    Write-Host "OK  $u -> $($r.StatusCode)"
  } catch {
    Write-Host "FAIL $u -> $($_.Exception.Message)"
  }
}
```

## 4. Open Platform Dashboard

- Dashboard UI: `http://localhost:5175`
- Unified API gateway: `http://localhost:8004`

The dashboard is built with `VITE_PLATFORM_API_URL=http://localhost:8004` in Compose.

## 4A. Auth Checklist For Protected Endpoints

Use this quick flow to validate role-based controls and webhook-key access.

1. Get a bearer token (worker login example):

```powershell
$loginBody = @{
  phone = "+919999999999"
  password = "demo-password"
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Method POST -Uri "http://localhost:8004/api/v1/auth/login" -ContentType "application/json" -Body $loginBody
$token = $loginResp.token
$authHeaders = @{ Authorization = "Bearer $token" }
```

1. Call a protected platform demo endpoint with bearer token:

```powershell
$demoBody = @{
  zone_id = 1
  force_order_drop = $true
  external_signal = "weather"
} | ConvertTo-Json

Invoke-RestMethod -Method POST -Uri "http://localhost:8004/api/v1/platform/demo/trigger-disruption" -Headers $authHeaders -ContentType "application/json" -Body $demoBody
```

1. Call a protected core internal endpoint with bearer token:

```powershell
Invoke-RestMethod -Method GET -Uri "http://localhost:8004/api/v1/internal/payouts/reconciliation" -Headers $authHeaders
```

1. Call a protected platform webhook using webhook key:

```powershell
$webhookHeaders = @{ "X-Platform-Webhook-Key" = $env:INDEL_PLATFORM_WEBHOOK_KEY }
$webhookBody = @{
  zone_id = 1
  source = "weather"
  status = "active"
} | ConvertTo-Json

Invoke-RestMethod -Method POST -Uri "http://localhost:8004/api/v1/platform/webhooks/external-signal" -Headers $webhookHeaders -ContentType "application/json" -Body $webhookBody
```

Expected behavior:

- Missing token/key: `401`
- Token/key present but insufficient role: `403`
- Valid role/key: `200`

Negative checks:

1. Missing token should fail with `401`:

```powershell
Invoke-WebRequest -Method POST -Uri "http://localhost:8004/api/v1/platform/demo/trigger-disruption" -ContentType "application/json" -Body '{"zone_id":1,"force_order_drop":true,"external_signal":"weather"}' -UseBasicParsing
```

1. Wrong webhook key should fail with `401` or `403`:

```powershell
$badWebhookHeaders = @{ "X-Platform-Webhook-Key" = "wrong-key" }
Invoke-WebRequest -Method POST -Uri "http://localhost:8004/api/v1/platform/webhooks/external-signal" -Headers $badWebhookHeaders -ContentType "application/json" -Body '{"zone_id":1,"source":"weather","status":"active"}' -UseBasicParsing
```

## 5. God Mode Routes

After opening the dashboard, use these pages:

- `/god-mode/temperature`
- `/god-mode/rain`
- `/god-mode/aqi`
- `/god-mode/traffic`
- `/god-mode/results`
- `/god-mode/batches`

## 6. API Routing Through Gateway

`api-gateway` forwards requests to the correct backend service:

- `/api/v1/platform/*` -> `platform-gateway:8003`
- `/api/v1/worker/*` -> `worker-gateway:8001`
- `/api/v1/auth/*` -> `worker-gateway:8001`
- `/api/v1/demo/*` -> `worker-gateway:8001`

## 7. Logs and Troubleshooting

View service logs:

```powershell
docker compose -f docker-compose.demo.yml logs -f api-gateway platform-dashboard worker-gateway platform-gateway
```

If you changed Docker or env values, rebuild:

```powershell
docker compose -f docker-compose.demo.yml up -d --build --force-recreate
```

If port conflicts occur, stop conflicting local services or update compose port mappings.

## 8. Recovery Procedure (Demo-Day Quick Path)

1. If one gateway is down:

```powershell
docker compose -f docker-compose.demo.yml restart worker-gateway insurer-gateway platform-gateway api-gateway
```

1. If API gateway is unhealthy:

```powershell
docker compose -f docker-compose.demo.yml restart api-gateway
docker compose -f docker-compose.demo.yml logs --tail=100 api-gateway
```

1. If one ML service is down (degraded mode possible):

```powershell
docker compose -f docker-compose.demo.yml restart premium-ml fraud-ml forecast-ml
```

1. If DB/migration order is broken:

```powershell
docker compose -f docker-compose.demo.yml up -d postgres db-migrate
docker compose -f docker-compose.demo.yml up -d core worker-gateway insurer-gateway platform-gateway api-gateway
```

1. Full reset fallback:

```powershell
docker compose -f docker-compose.demo.yml down -v
docker compose -f docker-compose.demo.yml up -d --build
```
