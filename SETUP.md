# Guide for Evaluation of Phase 3

This document is the evaluator runbook for **InDel — Insure, Deliver** (Guidewire DEVTrails 2026, Team ImaginAI).

It explains exactly what is needed and how to run the complete Phase 3 implementation locally: backend services, dashboards, and Android worker app.

---

## 1. Project Details

- Project: InDel — Insure, Deliver
- Team: ImaginAI
- Hackathon: Guidewire DEVTrails 2026
- Phase: Phase 3 (Scale & Optimize)

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

## 3. Repository Layout (What Runs in Phase 3)

- `backend/`: Go services (core, worker-gateway, insurer-gateway, platform-gateway, razorpay engine)
- `platform-dashboard/`: React + Vite dashboard for platform ops
- `insurer-dashboard/`: React + Vite dashboard for insurer ops
- `worker-app/`: Native Android app (Kotlin)
- `migrations/`: Database schema migrations
- `ml/`: Premium, fraud, and forecast model services
- `docker-compose.demo.yml`: Recommended compose orchestration for evaluation

---

## 4. Environment Setup
```
InDel/
│
├── .env                ← MASTER (Docker + backend + ML)
├── docker-compose.demo.yml
│
├── platform-dashboard/
│   └── .env           ← VITE_PLATFORM_API_URL
│
├── insurer-dashboard/
│   └── .env           ← VITE_INSURER_API_URL
│
├── worker-app/
│   └── .env           ← API_BASE_URL
```
Create a root `.env` file in the `InDel/` directory by copying the `.env.demo.example`.

Minimum values for local demo:

```env
# Network (Critical: Use your machine's LAN IP, e.g. 192.168.1.x, do NOT use localhost)
HOST_IP=<your_lan_ip>
LAN_IP=<your_lan_ip>
WORKER_API_URL=http://<your_lan_ip>:8004
PLATFORM_API_URL=http://<your_lan_ip>:8004
API_BASE_URL=http://<your_lan_ip>:8004

# Database
POSTGRES_PASSWORD=demo_password
DB_PASSWORD=demo_password

# NGINX Gateway & Vite Endpoints (Port 8004 reverse proxy)
VITE_PLATFORM_API_URL=http://<your_lan_ip>:8004
VITE_INSURER_API_URL=http://<your_lan_ip>:8004
VITE_CORE_API_URL=http://<your_lan_ip>:8004

# CORS Permissions
INDEL_ALLOWED_ORIGINS=http://<your_lan_ip>:5176,http://<your_lan_ip>:5175,http://<your_lan_ip>:5173

# Roles & Security Configuration
INDEL_DEMO_RESET_KEY=change-me-demo-reset-key
INDEL_DEMO_ALLOWED_ROLES=worker,admin,platform_admin,ops_manager
INDEL_DEMO_DESTRUCTIVE_ROLES=admin,platform_admin
INDEL_CORE_INTERNAL_ALLOWED_ROLES=worker,admin,platform_admin,ops_manager
INDEL_PLATFORM_OPERATOR_ALLOWED_ROLES=worker,admin,platform_admin,ops_manager
INDEL_PLATFORM_WEBHOOK_ALLOWED_ROLES=worker,admin,platform_admin,ops_manager
INDEL_PLATFORM_WEBHOOK_KEY=change-me-platform-webhook-key

# Razorpay Test Keys (Required for Live API Execution)
RAZORPAY_KEY_ID=<your_razorpay_key_id>
RAZORPAY_KEY_SECRET=<your_razorpay_key_secret>
```

Notes:

- If port binding errors happen on Windows with `127.0.0.1`, use your machine's local IPv4 (example: `192.168.x.x`).
- Do NOT commit your `.env` file to version control.

---

### 4.1 Payment Integration — Razorpay Setup

> [!NOTE]
> The backend executes real asynchronous transfers using the Razorpay API in Test Mode. The Go backend listens to the Kafka topic and triggers the payout automatically.

### Worker App — Secure Key Passing
We have eliminated hardcoded API keys from the Android application entirely.
You **do not** need to edit Kotlin files natively.

1. Ensure you have populated your Root `.env` file with your `RAZORPAY_KEY_ID`.
2. When you run Gradle Sync in Android Studio for the `worker-app`, the build configuration automatically pulls this `.env` variable securely into `BuildConfig.RAZORPAY_KEY_ID`.
3. If the app fails to open the Razorpay Checkout, simply confirm your `.env` is populated correctly and Rebuild the app.

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

> **Mandatory Setup for Evaluator:**
> Before running, ensure `/platform-dashboard/.env` contains your LAN IP:
> `VITE_PLATFORM_API_URL=http://<your_lan_ip>:8004`

Expected local URL (Vite):
- http://127.0.0.1:5173 (or next available port)

---

## 8. Run Insurer Dashboard

In a new terminal:

```bash
cd insurer-dashboard
npm install
npm run dev
```

> **Mandatory Setup for Evaluator:**
> Before running, ensure `/insurer-dashboard/.env` matches your LAN IP mapping exactly:
> `VITE_INSURER_API_URL=http://<your_lan_ip>:8004`
> `VITE_CORE_API_URL=http://<your_lan_ip>:8004`
> `VITE_PLATFORM_API_URL=http://<your_lan_ip>:8004`
> `VITE_ENABLE_API_DEBUG=true`

Expected local URL (Vite):
- http://127.0.0.1:5175 (or next available port)

---

## 9. Run Worker Mobile App (Android)

### Prerequisites

- Android Studio installed
- Android SDK and emulator installed
- Or physical Android device with USB debugging enabled

### Steps

1. Open `worker-app/` in Android Studio.
2. Wait for Gradle sync to complete.

> **Mandatory Setup for Evaluator:**
> Before running, ensure `/worker-app/.env` points to live gateways:
> `WORKER_API_URL=http://<your_lan_ip>:8004`
> `PLATFORM_API_URL=http://<your_lan_ip>:8004`

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
6. Trigger a disruption using the Chaos Engine — see Section 11 below.

---

## 11. Simulating a Disruption (Chaos Engine — Must Do)

> [!IMPORTANT]
> Nothing happens automatically. A disruption must be manually triggered
> to activate the full claims → fraud check → Kafka payout pipeline.

### Steps

1. Open the **Platform Dashboard** (http://127.0.0.1:5173)
2. Navigate to **Chaos Engine**
3. Click **"Collapse Demand"** to simulate a drop in order volume
4. Inject one or more disruption signals (at least one environmental signal is required):
   - Rain (Weather Alert)
   - AQI Spike
   - Zone Closure / Curfew

### What Happens Next

The system validates the disruption using multi-signal logic:

- Disruption is **confirmed** only when:
  - An environmental signal is present **AND**
  - Order volume drops significantly

Once confirmed, the automated pipeline kicks in:

1. Base claims are generated
2. Behavioral ML checks execute (`IsolationForest` / `DBSCAN`)
3. `auto_approved` claims publish to Kafka and execute via Razorpay
4. Worker wallet is updated in real time

---

### 11.1 Evaluating the 3-Layer Threat Engine (What to Look For)

### 11.1 Evaluating the Threat Engine (How to Test Fraud Scenarios)

> [!IMPORTANT]
>The system does NOT rely on pre-seeded mock claims to demonstrate fraud tracking. You must actively simulate the worker behaviors to prove the AI fraud engine intercepts malicious vectors. 

**Before** triggering the disruption (Step 11), simulate these behaviors using the Worker App or Batch Simulator:

1. **The Good Worker (Active & Insured)**: Open the Worker App, **pay the weekly premium via Razorpay**, and then take/complete an order normally. *(Use PIN: `1234` to verify the delivery).*
   → *Outcome:* After disruption, they show as `Paid` in the Insurer Dashboard. The system verified their active premium status and their legitimate order routing successfully.
2. **The Fraud Worker (Idle/Ghost)**: Log in, but sit completely idle and take **zero orders** before the disaster strikes.
   → *Outcome:* After disruption, they are intercepted into `Manual Review` natively. The ML and TTL gate catch the anomaly (e.g., `NO_LIVE_WINDOW_ACTIVITY`).
3. **The Uninsured/Offline Worker**: Log in but do NOT pay the premium (or remain completely offline).
   → *Outcome:* Completely ignored by the pipeline. They do not appear in the payout stream or the manual review queue at all, proving the policy gating works.

### 11.2 Verifying the Payout Execution

After the pipeline runs, confirm end-to-end by checking:

1. **Insurer Dashboard** → Claims stream — the auto-generated claim should show status `Paid` or `Manual Review`
2. **Worker App** → Wallet / Payouts — the credit entry will actively appear
3. **Backend logs** — you should see:
   - `Kafka consumer initiating Razorpay transfer...`
   - `Payout processed for worker...`

To check logs:
```bash
docker compose -f docker-compose.demo.yml logs -f core
```

---

## 12. Common Issues and Fixes

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

## 13. One-Command Quick Start (Evaluator Friendly)

From `InDel/`:

```bash
COMPOSE_PARALLEL_LIMIT=1 docker compose -f docker-compose.demo.yml up --build -d
```

Then run:

- `platform-dashboard` with `npm run dev`
- `insurer-dashboard` with `npm run dev`
- `worker-app` from Android Studio

---

**Guidewire DEVTrails 2026 — Phase 3 Submission**
