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
```
InDel/
│
├── .env                ← MASTER (Docker + backend + ML)
├── docker-compose.yml
│
├── platform-dashboard/
│   └── .env           ← VITE_PLATFORM_API_URL
│
├── insurer-dashboard/
│   └── .env           ← VITE_INSURER_API_URL
│
├── worker-app/
│   └── (hardcoded / config)
```
Create a root `.env` file in the `InDel/` directory.

Minimum values for local demo:

```env
# PostgreSQL
POSTGRES_USER=indel
POSTGRES_PASSWORD=<your_db_password>
POSTGRES_DB=indel
DB_USER=indel
DB_PASSWORD=<your_db_password>
DB_NAME=indel

# Network
HOST_IP=<your_local_ip>
API_BASE_URL=http://<your_local_ip>:8001/
API_BASE_URL1=http://<your_local_ip>:8003/

# ML Services (Local Access from Host)
PREMIUM_ML_URL=http://<your_local_ip>:9001
FRAUD_ML_URL=http://<your_local_ip>:9002
FORECAST_ML_URL=http://<your_local_ip>:9003

# Disruption APIs
OPENWEATHERMAP_API_KEY=<your_openweathermap_api_key>
OPENAQ_API_KEY=<your_openaq_api_key>

# Firebase
FIREBASE_API_KEY=<your_firebase_api_key>
FIREBASE_PROJECT_ID=<your_project_id>
FIREBASE_PROJECT_NUMBER=<your_project_number>
FIREBASE_STORAGE_BUCKET=<your_storage_bucket>
FIREBASE_APP_ID=<your_app_id>
FIREBASE_SERVER_KEY=<your_server_key>

# Kafka
KAFKA_BROKERS=kafka:9092

# Insurer Dashboard
VITE_INSURER_API_URL=http://<your_local_ip>:8002
VITE_CORE_API_URL=http://<your_local_ip>:8000

# Platform Dashboard
VITE_PLATFORM_API_URL=http://<your_local_ip>:8003

# Razorpay Test Keys
TEST_KEY_ID=<your_razorpay_test_key_id>
TEST_KEY_SECRET=<your_razorpay_test_key_secret>
```

Notes:

- On Windows, if port binding errors happen with `HOST_IP=127.0.0.1`, use your machine local IPv4 (example: `192.168.x.x`).
---

### 4.1 Payment Integration — Demo Mode (Minimal Setup)

> [!NOTE]
> Razorpay runs in **Mock Mode** by default for backend payouts.

- All financial transactions are automatically simulated
- The backend detects missing keys and switches to mock mode seamlessly
- Payout IDs in dashboards will appear as `rzp_mock_...`
- Worker wallets update in real time — identical behavior to a live environment

> [!IMPORTANT]
> The backend runs fully in Mock Mode and requires no Razorpay credentials.
> However, the Worker App uses the Razorpay Android SDK for payment UI rendering,
> which requires a valid test key to load correctly.

### Worker App — One-Time Key Setup

1. Open this file in Android Studio:
   `worker-app/app/src/main/java/com/imaginai/indel/MainActivity.kt`

2. Find this line (around line 41):
```kotlin
   val razorpayKeyId = "rzp_test_REPLACE_WITH_YOUR_KEY"
```

3. Replace the placeholder with your Razorpay **test** key ID from the [Razorpay Dashboard](https://dashboard.razorpay.com/).

> Only the payment UI rendering requires this key. All actual payout logic and
> verification runs through the mock backend — no real money is involved.

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
6. Trigger a disruption using the Chaos Engine — see Section 11 below.

---

## 11. Simulating a Disruption (Chaos Engine — Must Do)

> [!IMPORTANT]
> Nothing happens automatically. A disruption must be manually triggered
> to activate the full claims → fraud check → payout pipeline.

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

1. Claims are generated
2. Fraud check runs
3. Payout is processed via Kafka
4. Worker wallet is updated in real time

You can verify each stage on the **Insurer Dashboard** and in the **Worker App**.

### Bonus: Dynamic Premium Behavior

Trigger disruptions multiple times to observe adaptive risk scoring:

- Risk score updates after each disruption
- Future premiums increase based on zone instability
- This reflects the ML-driven premium adjustment model in action
  

### Verifying the Payout (What to Look For)

After the pipeline runs, confirm end-to-end by checking:

1. **Insurer Dashboard** → Claims — the auto-generated claim should show status `Paid`
2. **Worker App** → Wallet / Payouts — a credit entry with a mock transaction ID (`rzp_mock_...`) will appear
3. **Backend logs** — you should see:
   - `Razorpay client initialized (Mock Mode: true)`
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

## 14. Scope Note

This guide is focused on **Phase 2 implementation evaluation**. It prioritizes reproducible local execution and validation over production deployment concerns.
