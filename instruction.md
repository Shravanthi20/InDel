# Backend Run Instructions (InDel)

## 1. Prerequisites

- Docker Desktop running
- Docker Compose v2 (`docker compose`)
- Ports free: `5432`, `8000`, `8001`, `8002`, `8003`, `9092`, `2181`

## 2. Configure environment

From project root, create `.env` from `.env.example` if you do not already have one:

```powershell
Copy-Item .env.example .env
```

Set these values in `.env` (important):

- `HOST_IP=127.0.0.1` for same-machine testing
- `HOST_IP=<your LAN IP>` (example `192.168.1.6`) for phone/device testing
- `DB_PASSWORD` and `POSTGRES_PASSWORD` should match the compose defaults you want to use

## 3. Start backend stack (demo)

From project root:

```powershell
docker compose -f docker-compose.demo.yml down -v
docker compose -f docker-compose.demo.yml up -d --build
```

Check containers:

```powershell
docker compose -f docker-compose.demo.yml ps
```

## 4. Verify backend is up

Quick health-style checks:

```powershell
Invoke-WebRequest "http://localhost:8001/api/v1/demo/orders/available?limit=1" -UseBasicParsing
Invoke-WebRequest "http://localhost:8001/api/v1/worker/batches?limit=1" -UseBasicParsing
```

If `HOST_IP` is LAN IP, also verify with that IP from your device/browser.

## 5. Seed demo orders (optional but recommended)

```powershell
python scripts/fake_order_publisher.py --generate --orders-per-run 2 --reset-url http://localhost:8001/api/v1/demo/reset --url http://localhost:8001/api/v1/demo/orders/ingest
```

For LAN/device flow, replace `localhost` with your LAN IP.

## 6. Useful logs

Tail gateway logs:

```powershell
docker compose -f docker-compose.demo.yml logs -f worker-gateway
docker compose -f docker-compose.demo.yml logs -f core
```

## 7. Stop backend stack

```powershell
docker compose -f docker-compose.demo.yml down
```

Remove containers + volumes (full reset):

```powershell
docker compose -f docker-compose.demo.yml down -v
```
