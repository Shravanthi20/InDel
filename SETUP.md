# Guide for Evaluation of Phase 2

This document is the evaluator runbook for **InDel — Insure, Deliver** (Guidewire DEVTrails 2026, Team ImaginAI).

It explains exactly what is needed and how to run the complete Phase 2 implementation locally: backend services, dashboards, and Android worker app.

> [!IMPORTANT]
> For Phase 2 evaluation, we are not distributing a standalone APK. We kindly request evaluators to run the Worker App using the provided project setup (Android Studio + local backend services) so the full integrated workflow can be verified.

---

## 1. Project Details

- Project: InDel — Insure, Deliver
- Team: ImaginAI
- Hackathon: Guidewire DEVTrails 2026
- Phase: Phase 2 (Implementation)

### Team Members

- Shravanthi S: Core Policy, Premium Cycle, Payout and Data Operations
- Gayathri U: Delivery Management and DevOps
- Rithanya K A: ML Services (Training and Serving)
- Saravana Priyaa C R: Platform Integration, Disruption Engine
- Subikha MV: Insurer System, Claims Intelligence and System Design

---

## 2. What You Need Installed

Install all prerequisites before running anything.

### Mandatory

- Git
- Docker Desktop (must be running)
- Docker Compose (included with Docker Desktop)
- Node.js 18+ and npm
- Android Studio (for running the worker app)

### Optional but useful

- Postman or similar API client
- A modern browser (Chrome/Edge)

---

## 3. Repository Layout (What Runs in Phase 2)

- `backend/`: Go services (core, worker-gateway, insurer-gateway, platform-gateway)
- `platform-dashboard/`: React + Vite dashboard for platform ops
- `insurer-dashboard/`: React + Vite dashboard for insurer ops
- `worker-app/`: Native Android app (Kotlin)
- `migrations/`: Database schema migrations
- `ml/`: Premium, fraud, and forecast model services
- `docker-compose.demo.yml`: Recommended compose file for evaluation/demo

---

## 4. Environment Setup

Create a root `.env` file in the `InDel/` directory.

Minimum values for local demo:

```env
HOST_IP=127.0.0.1
DB_PASSWORD=demo_password
POSTGRES_PASSWORD=demo_password
```

Notes:

- On Windows, if port binding errors happen with `HOST_IP=127.0.0.1`, use your machine local IPv4 (example: `192.168.x.x`).
- Do not commit real secrets in `.env`.

---

## 5. Start Full Backend Stack (Recommended for Evaluation)

From repository root (`InDel/`), run:

```bash
COMPOSE_PARALLEL_LIMIT=1 docker compose -f docker-compose.demo.yml up --build -d
```

This starts:

- PostgreSQL
- Migration runner
- Kafka + Zookeeper
- core service (8000)
- worker-gateway (8001)
- insurer-gateway (8002)
- platform-gateway (8003)
- ML services:
  - premium-ml (9001)
  - fraud-ml (9002)
  - forecast-ml (9003)

To stop:

```bash
docker compose -f docker-compose.demo.yml down
```

To stop and remove volumes (fresh DB reset):

```bash
docker compose -f docker-compose.demo.yml down -v
```

---

## 6. Health Checks (Must Pass Before UI Testing)

Open these URLs in browser:

- http://127.0.0.1:8000/health
- http://127.0.0.1:8001/health
- http://127.0.0.1:8002/health
- http://127.0.0.1:8003/health

If a service is down, inspect logs:

```bash
docker compose -f docker-compose.demo.yml logs -f core
```

Replace `core` with `worker-gateway`, `insurer-gateway`, `platform-gateway`, etc.

---

## 7. Run Platform Dashboard

In a new terminal:

```bash
cd platform-dashboard
npm install
npm run dev
```

Expected local URL (Vite):

- http://127.0.0.1:5173 (or next available port shown in terminal)

If required, set API endpoint in `platform-dashboard/.env`:

```env
VITE_PLATFORM_API_URL=http://127.0.0.1:8003
```

---

## 8. Run Insurer Dashboard

In a new terminal:

```bash
cd insurer-dashboard
npm install
npm run dev
```

Expected local URL (Vite):

- http://127.0.0.1:5174 (or next available port shown in terminal)

If required, set API endpoints in `insurer-dashboard/.env`:

```env
VITE_INSURER_API_URL=http://127.0.0.1:8002
VITE_CORE_API_URL=http://127.0.0.1:8000
```

---

## 9. Run Worker Mobile App (Android)

### Prerequisites

- Android Studio installed
- Android SDK and emulator installed
- Or physical Android device with USB debugging enabled

### Steps

1. Open `worker-app/` in Android Studio.
2. Wait for Gradle sync to complete.
3. Select emulator/device.
4. Run the app.

### Networking note for physical device

If app APIs use localhost, replace backend host with your machine LAN IP (same Wi-Fi for laptop and phone).

---

## 10. Evaluation Flow (Suggested Demo Path)

1. Start all Docker services with `docker-compose.demo.yml`.
2. Verify health endpoints (all 4 gateways).
3. Open Platform Dashboard and verify operations pages load.
4. Open Insurer Dashboard and verify risk/claims pages load.
5. Run Worker App and verify it can hit backend APIs.
6. Trigger a disruption flow from platform/admin endpoints and observe claim/payout pipeline updates.

---

## 11. Common Issues and Fixes

### Docker build/start fails

- Ensure Docker Desktop is running.
- Retry with:

```bash
docker compose -f docker-compose.demo.yml down
docker compose -f docker-compose.demo.yml up --build -d
```

### Port already in use

- Free the port or stop conflicting process/service.
- Check ports: 5432, 2181, 9092, 8000-8003, 9001-9003.

### Dashboard cannot reach backend

- Verify backend health endpoints first.
- Verify Vite `VITE_*` URLs point to correct local gateway ports.

### Android app cannot connect

- Use machine LAN IP, not `127.0.0.1`, when running on physical device.
- Ensure device and laptop are on same network.

---

## 12. One-Command Quick Start (Evaluator Friendly)

From `InDel/`:

```bash
COMPOSE_PARALLEL_LIMIT=1 docker compose -f docker-compose.demo.yml up --build -d
```

Then run:

- `platform-dashboard` with `npm run dev`
- `insurer-dashboard` with `npm run dev`
- `worker-app` from Android Studio

---

## 13. Scope Note

This guide is focused on **Phase 2 implementation evaluation**. It prioritizes reproducible local execution and validation over production deployment concerns.

