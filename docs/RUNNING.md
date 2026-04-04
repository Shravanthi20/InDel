# Running InDel Locally

This guide covers the updated stack with:

- backend gateways (`core`, `worker-gateway`, `insurer-gateway`, `platform-gateway`)
- ML services (`premium-ml`, `fraud-ml`, `forecast-ml`)
- unified API gateway (`api-gateway` on `:8004`)
- Platform Dashboard (`platform-dashboard` on `:5174`)

## Prerequisites

- Docker Desktop (with Compose v2)
- Ports available: `5174`, `8000`, `8001`, `8002`, `8003`, `8004`, `9001`, `9002`, `9003`

## 1. Start Demo Stack (Recommended)

From repository root:

```powershell
docker compose -f docker-compose.demo.yml down -v
docker compose -f docker-compose.demo.yml up -d --build
```

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
  'http://localhost:8001/health',
  'http://localhost:8002/health',
  'http://localhost:8003/health',
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

- Dashboard UI: `http://localhost:5174`
- Unified API gateway: `http://localhost:8004`

The dashboard is built with `VITE_PLATFORM_API_URL=http://localhost:8004` in Compose.

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
