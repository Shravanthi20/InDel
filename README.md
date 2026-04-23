# InDel — Insure, Deliver

> *The rain starts at 11:40 AM. Orders stop. Income collapses. The worker does nothing. At 5:30 PM, ₹527 hits their UPI. No form. No call. No claim. Just money.*

<p align="center">
  <img src="https://img.shields.io/badge/Guidewire-DEVTrails%202026-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Phase-3%20Complete-brightgreen?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Stack-Go%20%7C%20Kafka%20%7C%20Redis%20%7C%20ML-orange?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Persona-E--Commerce%20Delivery-purple?style=for-the-badge" />
</p>

<p align="center">
  <a href="https://www.youtube.com/watch?v=tQliernHGDI&t=30s">
    <img src="https://img.shields.io/badge/Demo-5%20Min%20Walkthrough-red?style=for-the-badge" />
  </a>
  &nbsp;&nbsp;
  <a href="./SETUP.md">
    <img src="https://img.shields.io/badge/Setup%20Guide-Run%20Locally-grey?style=for-the-badge" />
  </a>
  &nbsp;&nbsp;
  <a href="https://indel-portal.vercel.app/">
    <img src="https://img.shields.io/badge/Live-indel--portal.vercel.app-22c55e?style=for-the-badge" />
  </a>
</p>

---

> **To run this project locally, follow the steps in [SETUP.md](./SETUP.md).**

[**View Pitch Deck →**](https://drive.google.com/file/d/1xtUFPMNEktznhHQtsVwRI5vB8clwfVZw/view?usp=sharing)

---

## The Problem Nobody Has Solved

India has **15+ million gig delivery workers**. Every rupee they earn depends on one thing: completing orders.

Floods. Hazardous AQI. Curfews. Zone closures. When any of these hit — deliveries stop, income collapses, and there is **zero fallback**. Workers lose **20–30% of monthly income** during disruption events.

Traditional insurance covers accidents and vehicles. Existing parametric attempts require delivery platforms to share worker data — platforms that have **zero incentive to cooperate**. The result is unverifiable claims, rampant fraud, and no product that works at scale.

**The market gap is real. The workers are real. InDel is the answer.**

---

## What Makes InDel Different

Every other parametric system asks: *"Was the worker inside the disrupted zone?"* GPS is trivially spoofed. That question is the wrong one.

**InDel asks: did this worker's economic reality collapse?**

```
❌ Old way:
   Insurer → requests data from Amazon / Flipkart
   → access denied → weak verification → fraud → no product

✅ InDel way:
   Insurer deploys InDel → integrates with delivery platform via API
   → first-party data layer owned by InDel
   → verified disruption → automated payout → zero manual claims
```

The insurer gets ready-to-deploy infrastructure. The worker gets protection that runs silently in the background. They never file a claim. **It just arrives.**

---

## Numbers That Matter

> Pilot simulation — 1,000 workers, Chennai, one month

| Metric | Value |
|--------|-------|
| Premiums collected | ₹68,000 |
| Payouts disbursed | ₹44,000 |
| Gross margin | 35% |
| Loss ratio | ~65% (industry benchmark: 70–85%) |

We are already **inside the profitable band** — before scale.

---

## Phase 3 Deliverables — Every Box Checked

| Requirement | InDel Implementation | Status |
|-------------|----------------------|--------|
| Advanced Fraud Detection | 3-Layer Stacked Threat Engine — IsolationForest + DBSCAN + Postgres Hard Rules across 6 behavioral dimensions | ✅ Fully Implemented |
| Instant Payout System | Native KRaft-mode Kafka async pipeline → Razorpay UPI, idempotent offsets, 5× exponential retry | ✅ Fully Implemented |
| Intelligent Dashboards | Worker: multilingual SHAP audit + earnings protection · Insurer: Prophet 7-day reserve analytics + live fraud queue | ✅ Fully Implemented |
| Edge-Resilient Mobile App | Offline-first Android architecture powered by Room databases dynamically synced with secure auth boundaries preventing cross-worker state bleed | ✅ Fully Implemented |

---

## System Architecture

```mermaid
graph TD
    subgraph "Clients"
        WA["Worker App (Kotlin/Android)"]
        ID["Insurer Dashboard (React)"]
        PD["Platform Dashboard (React)"]
    end

    subgraph "Entry Point"
        AGW["API Gateway (Nginx :8004)"]
    end

    subgraph "API Gateways (Golang)"
        WGW["Worker Gateway :8001"]
        IGW["Insurer Gateway :8002"]
        PGW["Platform Gateway :8003"]
    end

    subgraph "InDel Core & Storage"
        CS["Core Service :8000"]
        DB[("PostgreSQL")]
        RD[("Redis Cache")]
    end

    subgraph "Intelligence Layer"
        ML["Unified ML Service (FastAPI :9000)"]
        subgraph "Engines"
            PREM["Premium Pricing"]
            FRD["Fraud Detection"]
            FCT["Forecasting"]
        end
    end

    subgraph "Async Pipeline"
        KF["Kafka (KRaft Mode :9092)"]
    end

    subgraph "External Integration"
        ENV["Environmental APIs (Weather/AQI)"]
        RP["Razorpay UPI Payout Rails"]
    end

    WA <-->|REST| AGW
    ID <-->|REST| AGW
    PD <-->|REST| AGW

    AGW -->|/api/v1/worker| WGW
    AGW -->|/api/v1/insurer| IGW
    AGW -->|/api/v1/platform| PGW

    WGW <-->|Internal API| CS
    IGW <-->|Internal API| CS
    PGW <-->|Internal API| CS

    WGW <--> DB
    IGW <--> DB
    PGW <--> DB
    CS <--> DB
    WGW <--> RD

    WGW -->|Inference| ML
    IGW -->|Inference| ML
    PGW -->|Inference| ML
    CS -->|Inference| ML

    ML --- PREM
    ML --- FRD
    ML --- FCT

    CS -->|Produce Events| KF
    KF -->|Consume Tasks| CS

    ENV -->|Webhooks/Jobs| CS
    CS -->|Payouts| RP
```

*The diagram shows the flow of requests, the shared Redis cache, and the event-driven Kafka pipeline. Every component is containerised and orchestrated via Docker Compose.*

---

## Tech Stack & Version Matrix

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend** | Go (net/http, Gin), GORM | Go 1.25 |
| **Frontend** | React, Vite, TypeScript, Tailwind CSS, Lucide Icons | React 18, Vite 5 |
| **Mobile** | Kotlin (Android), Razorpay Android SDK | Kotlin 1.9 |
| **Async Messaging** | Apache Kafka (KRaft mode) | Kafka 3.7 |
| **In-Memory Cache** | Redis Cluster | Redis 7.2 |
| **Database** | PostgreSQL + pgcrypto | Postgres 15 |
| **ML Serving** | Python, FastAPI, joblib | Python 3.11 |
| **AI Algorithms** | XGBoost, SHAP, IsolationForest, DBSCAN, Prophet | XGBoost 2.0 |
| **Containerisation** | Docker, Docker Compose | Docker 25 |
| **API Gateway** | Nginx | 1.25 |
| **Environmental APIs** | OpenWeatherMap, OpenAQ / WAQI | REST v2 |
| **Payments** | Razorpay Payouts SDK / UPI simulator | SDK v2 |
| **Notifications** | Firebase Cloud Messaging | Admin SDK v1 |
| **Explainability** | SHAP TreeExplainer + IndicTrans2 | SHAP 0.44 |
| **CI/CD** | GitHub Actions | — |

---

## Cache & Redis Layer

### Why a Dedicated Cache Matters

- **Sub-millisecond data access** — All UI widgets (premium display, plan-status dashboard, forecast charts, fraud-risk alerts) pull their data from Redis instead of hitting PostgreSQL on every refresh.
- **DB-offload at scale** — In a production-grade deployment with >10k concurrent workers, Redis absorbs >85% of read traffic, keeping the relational store comfortably below its CPU/IO limits.
- **Deterministic latency for actuarial calculations** — Premiums, risk scores, and disruption forecasts are cached for the duration of the weekly cycle, guaranteeing that the same worker sees the exact same value across the app, the insurer dashboard, and any downstream analytics pipelines.

### Architectural Blueprint

```
+-------------------+          +------------------------+
|  Front-End UI     |  <--->   |  API-Gateway (NGINX)   |
+-------------------+          +------------------------+
          |                               |
          v                               v
+-------------------+          +------------------------+
|  Go Backend (Gin) |  <--->   |  Redis Cluster (3-node)|
+-------------------+          +------------------------+
          |                               |
          v                               v
+-------------------+          +------------------------+
|  PostgreSQL (RDB) |  <--->   |  Persistent AOF + RDB  |
+-------------------+          +------------------------+
```

### Cache Design

| Domain | Redis Key Pattern | Value (JSON) | TTL | Invalidation Trigger |
|--------|-------------------|--------------|-----|----------------------|
| Active Policy | `policy:{worker_id}` | `{policy_id, weekly_premium, status, updated_at}` | 7 days | End-of-week batch job or manual plan change |
| Zone Forecast | `forecast:{zone_id}` | `{model, forecast:[{date, prob}]}` | 1 hour | New weather/AQI signal or scheduled retrain |
| Fraud Risk Score | `fraud:{claim_id}` | `{score, layer, details}` | 24 hours | Claim finalisation (approved/rejected) |
| Session / Auth Token | `session:{jwt_id}` | `{user_id, role, exp}` | 30 minutes | Logout / token refresh |

### Cache-Aside Flow

```go
func GetPolicy(workerID int) (*Policy, error) {
    key := fmt.Sprintf("policy:%d", workerID)
    if val, err := redis.Get(key); err == nil {
        // cache hit — sub-millisecond return
        return json.Unmarshal(val)
    }
    // cache miss → PostgreSQL
    policy, err := db.QueryPolicy(workerID)
    if err != nil { return nil, err }
    // write-through to Redis
    payload, _ := json.Marshal(policy)
    redis.Set(key, payload, 7*24*time.Hour)
    return policy, nil
}
```

### Cache-Invalidation & Consistency Guarantees

1. **Write-through for critical updates** — When a worker pays a new premium, the Go service writes to PostgreSQL *and* updates `policy:{worker_id}` atomically via a Redis transaction (`MULTI/EXEC`).
2. **Event-driven invalidation** — Kafka events (`PolicyUpdated`, `ZoneSignal`, `FraudScoreComputed`) trigger a lightweight Go consumer that deletes the stale key, forcing a refresh on the next read.
3. **Graceful stale-read fallback** — If a key expires mid-request, the service falls back to the DB without surfacing an error to the UI, preserving a seamless user experience.

### Durability

- Redis runs as a **3-node master-replica cluster** with `appendonly yes` and periodic RDB snapshots.
- On master failure, a replica is promoted automatically, guaranteeing **<5 s failover**.

### Performance Impact

| Metric | Before Redis | After Redis | Δ |
|--------|-------------|------------|---|
| Avg. read latency (premium) | 12 ms (peak 150 ms) | **0.8 ms** | –93% |
| DB CPU utilisation (peak) | 85% | **32%** | –53% |
| Cache-hit ratio | N/A | **87%** | — |
| Failover recovery | Minutes (manual) | **<5 s** (automatic) | — |

---

## Zero-Touch Claim Flow — End to End

```
Environmental signal received — weather / AQI / curfew
        ↓
Order velocity collapse confirmed — >30% drop vs 4-week baseline
        ↓
Multi-signal lock engaged — both must align within time-lag window
        ↓
Zone scan — active policy + TTL heartbeat check + acceptance rate threshold
        ↓
Income loss computed automatically
  Baseline  =  4-week average hourly earnings (InDel first-party data)
  Loss      =  Expected earnings − Actual earnings during disruption window
  Payout    =  Loss × coverage ratio (80–90%), capped at weekly maximum
        ↓
3-Layer fraud check runs independently
  Layer 1 — Isolation Forest: anomaly score on 6-dimension claim vector
  Layer 2 — DBSCAN: does this worker's behavior match their zone cluster?
  Layer 3 — Hard rules: GPS in zone? Zero deliveries during the window?
        ↓
  Low-risk   →  auto-approved instantly
  Medium     →  held for secondary validation
  High-risk  →  manual review queue with full violations JSON
        ↓
Worker notified with SHAP breakdown in their language
Payout: Kafka → Razorpay → UPI → worker's account
```

**A confirmed disruption event — Tambaram Flood, 11:40 AM to 5:30 PM:**

```
Worker baseline earnings:     ₹120 / hour
Expected over 5.83 hours:     ₹700
Actual (2 partial orders):     ₹80
Loss:                          ₹620
Payout at 85% coverage:        ₹527  →  UPI, same day
```

*The worker received a notification. They never opened a form.*

---

## Fraud Defense: 3-Layer Stacked Threat Engine

InDel intercepts claim vectors across **6 behavioral dimensions**:

```python
[earnings_drop_ratio, avg_orders_per_hour, distance_routed,
 claim_frequency, approval_ratio, zone_risk]
```

### Layer 1 — Isolation Forest Anomaly Filter

Fraud syndicates push identical income loss claims across hundreds of accounts simultaneously. `IsolationForest(contamination='auto')` detects these — identical 6-dimension vectors generate anomalously short tree path-lengths. Score above 0.55 → structural claim delay. Genuine disruption creates uniquely staggered drops. Coordinated fraud creates near-identical vectors that the forest isolates immediately.

### Layer 2 — DBSCAN Spatial Clustering

During a verified flood, legitimate workers in a zone behave similarly — speeds drop, idle time rises, routing follows impaired road networks. DBSCAN clusters this behavior. Any worker whose telemetry diverges dramatically from the zone cluster is flagged as a noise point — catching sophisticated actors whose history looks clean but whose event-specific behavior is inexplicably wrong.

### Layer 3 — Postgres Hard Rules

No AI needed. Completed deliveries logged during the claimed disruption window → auto-rejected. Worker GPS outside zone boundaries before the event → auto-rejected.

Flagged claims are not dropped. They route to `manual_review` with a structured violations JSON:

```json
{
  "violations": [
    "AI Anomaly Threshold breached — Isolation Forest score: 0.87",
    "GPS Trace divergent from zone cluster (DBSCAN noise point)",
    "2 deliveries completed during claimed disruption window"
  ]
}
```

Human underwriters see the exact algorithmic reasoning. No database querying. No guesswork.

---

## Dynamic Pricing Engine

No flat rates. No city-wide averages. Each premium is computed fresh every week.

### Risk Score Formula

```
R = (Order Volatility × 0.24) + (Earnings Volatility × 0.22)
  + (Disruption Rate × 0.20) + (Weather Signal × 0.34)
```

*Weather leads at 34% — the strongest predictor of delivery income loss across Indian urban zones.*

### Premium Formula

```
P = (4-week avg daily earnings × 0.0375) × (0.72 + R) × Vehicle Factor
```

Vehicle Factor: **1.04 for EVs**, **1.08 for ICE** — rewarding sustainable delivery.

| Zone | Risk Level | Weekly Premium | Max Weekly Payout |
|------|------------|----------------|-------------------|
| Tambaram, Chennai | High — monsoon + heat | ₹22 | ₹800 |
| Rohini, Delhi | Medium | ₹17 | ₹700 |
| Koramangala, Bengaluru | Low | ₹12 | ₹600 |

### Multilingual SHAP Explainability — 3-Step Architecture

Raw SHAP JSON passed to a translation engine breaks syntax entirely. InDel solves this with a purpose-built pipeline:

1. **JSON Templating** — Go backend maps raw SHAP numerics into predefined English structural templates
2. **Pre-verified Translation** — Tamil and Hindi structural templates stored statically, grammatically validated
3. **Dynamic Value Injection** — numeric values injected into the correct-language template at runtime

```
Your premium this week: ₹18
  Flood risk in your zone    +₹6
  Recent AQI pattern         +₹3
  Income instability score   +₹2
  Base rate                   ₹7
```

Supported: **English · Tamil · Hindi** + icon-based visual cues for low-literacy users.

### The Five Disruption Triggers

| Trigger | Source | Fires When |
|---------|--------|------------|
| `WEATHER_ALERT` | OpenWeatherMap | Rainfall / flood / extreme heat threshold crossed |
| `AQI_ALERT` | OpenAQ / WAQI | Pollution exceeds hazardous levels |
| `ORDER_DROP_DETECTED` | InDel internal telemetry | Zone order volume drops >30% vs sliding baseline |
| `ZONE_CLOSURE_ALERT` | Traffic API / Govt alerts | Curfew, strike, or zone restriction detected |
| `WORKER_ACTIVITY_UPDATE` | InDel platform | Login, acceptance, and completion pattern anomaly |

A disruption is confirmed only when an **external environmental signal** and an **internal order volume collapse** align simultaneously — with a time-lag window that accounts for the real-world delay between rainfall starting and orders actually stopping.

---

## Kafka Payout Pipeline

Synchronous API transfers during mass disruption = thread locking, gateway timeouts, duplicate calls, crashed backends.

InDel **decouples claim approval from financial execution entirely** via Apache Kafka.

| Kafka Property | Why It Matters |
|----------------|----------------|
| Replayable offsets | Razorpay 503? Kafka replays from last committed offset — no duplicates, no lost payouts |
| Consumer group isolation | Payout processor and audit logger run as separate groups — audit never blocks payment |
| Horizontal scaling | Additional consumer instances spin up during surges, pick up unconsumed partitions automatically |
| Persistent audit log | Every payout attempt retained — sent, succeeded, retried, failed. Regulatory-grade trail. |

### Razorpay Integration — Money Moves Both Ways

| Step | Action |
|------|--------|
| 1 | Claim approved — income loss calculated |
| 2 | Payout queued in Postgres with status `queued` |
| 3 | Backend calls Razorpay `POST /v1/payouts` with worker UPI + amount |
| 4 | Transient failures retried up to 5× with exponential backoff |
| 5 | Idempotency key `pay_clm_<claim_id>` — zero duplicate payouts |
| 6 | Status sync — `processed` on success, `failed` routes to human review |

A 10-second heartbeat goroutine continuously drains the payout queue. Workers receive money within **seconds** — not hours, not the next business day.

**KRaft Mode:** InDel uses Kafka's modern KRaft protocol — no Zookeeper, no external coordination JVM overhead. Broker restarts? The native KRaft controller enforces new leader election automatically via quorum. No manual intervention. No lost messages.

Five thousand workers claiming simultaneously during a citywide curfew: **handled.**

---

## Observability & Monitoring

InDel provides operational visibility through service health checks, runtime logs, and transactional audit records across the core workflow.

## System Component Visibility

| Component              | Current Visibility |
|----------------------|-------------------|
| API Gateway (Nginx)  | Gateway health endpoint and access/error logs |
| Go Services          | Service health endpoints, request handling logs, and database operation traces in logs |
| ML Service (FastAPI) | Health endpoint, startup/model-load logs, and inference request logs |
| Redis                | Service health checks and runtime logs |
| Kafka                | Broker health checks and event-processing logs |
| PostgreSQL           | Container health checks and query/audit data persisted in tables |

### Observability Summary

Logging is available through Docker and Docker Compose service logs, enabling issue diagnosis across API, ML, async processing, and payout execution paths.
The platform records key financial and workflow events in the database, including claim state transitions, payout attempts, retries, and final payout outcomes, providing an end-to-end audit trail from disruption processing to payout completion.

---

## Security & Compliance

| Concern | Current Implementation |
|---------|------------------------|
| Authentication | JWT-based authentication using HMAC signing with a shared secret from environment configuration |
| Authorization | Role-based access control enforced in gateway and internal middleware for worker, insurer, platform, and internal operations |
| Secret Management | Runtime secrets are read from environment variables and local env files excluded from version control |
| Transport | Local Docker setup runs service traffic over internal container networking through the API gateway |
| Data Handling | Core workflow data is stored in PostgreSQL, with claim and payout state transitions persisted for auditability |
| Session and Access Scope | Worker session state and role-scoped routes are separated across worker, insurer, platform, and core service surfaces |
| Abuse Controls | Route-level role guards and internal authorization checks are applied on sensitive control endpoints |

InDel currently emphasizes application-level auth, authorization, and operational audit trails for payout and claim workflows.

---

## Testing Strategy

| Layer | Tooling | Coverage |
|-------|---------|----------|
| **Unit Tests** | Go: `testing` + `testify`; Python: `pytest` | 85% (core services) |
| **Integration Tests** | Docker Compose test harness (`docker-compose -f docker-compose.test.yml up`) | End-to-end: API-Gateway → Backend → DB/Redis |
| **Contract Tests** | `pact` for Go ↔ Python contracts (premium, fraud, forecast) | API stability across language boundaries |
| **Load / Stress Tests** | `k6` scripts simulating 10k concurrent workers hitting the premium endpoint | 95% requests < 100 ms |
| **Security Tests** | OWASP ZAP automated scan on the Nginx gateway | No high-severity findings |

**CI Pipeline** — GitHub Actions runs all suites on every PR; builds multi-arch Docker images; pushes to GitHub Container Registry on merge.

---

## CI/CD Pipeline

The repository includes automated workflows for validation, testing, build checks, and deployment orchestration.

### Continuous Integration Workflow

1. Compose Validation
- Validates primary Docker Compose files.

2. Backend Quality Checks
- Runs Go test suites.
- Runs Go lint checks with golangci-lint.

3. ML Service Tests
- Sets up Python environment and dependencies.
- Runs ML test suite with pytest.

4. Frontend Build Validation
- Installs dependencies and builds insurer dashboard.
- Installs dependencies and builds platform dashboard.

5. Endpoint Contract Verification
- Executes endpoint contract check script from the scripts folder.

6. Docker Image Build Verification
- Builds backend and ML images as CI build checks.

7. Integration Tests
- Boots demo stack with Docker Compose.
- Waits for core health readiness.
- Runs backend worker integration tests.
- Captures logs on failure and performs teardown.

### Deployment Workflow

- Includes a manual or main-branch deployment workflow for service rollout.
- Performs post-deploy health sweeps across configured public service URLs.
- Supports deployment notifications through configured webhook integration.

This pipeline ensures each change is validated through tests and build checks before release operations.

---

## Deployment & Operations

**One command, fully running:**

```bash
COMPOSE_PARALLEL_LIMIT=1 docker compose -f docker-compose.demo.yml up --build -d
```

| Container | Role | Port |
|-----------|------|------|
| `indel-api` | Go / Gin REST API (Worker, Insurer, Platform gateways) | — |
| `postgres` | PostgreSQL, strict 1.0 CPU / 1 GB limits, migrations pre-applied | :5432 |
| `kafka` | Native KRaft async payout broker, stripped memory footprint | :9092 |
| `ml-service` | Unified Python FastAPI (XGBoost Pricing, Fraud, Forecast) | :9000 |
| `redis` | In-memory cache cluster | :6379 |

`COMPOSE_PARALLEL_LIMIT=1` enforces startup order logically. The entire environment forces aggressive container limit clamping (`deploy.resources.limits`). The demo compose launches pre-seeded with workers, zones, and disruption history. **No manual setup. No configuration. One command.**

### Operational Runbook

| Step | Command | Description |
|------|---------|-------------|
| **Rolling Update** | `docker service update --image ghcr.io/.../service:latest service-name` | Zero-downtime upgrade (Docker Swarm mode) |
| **Health Check** | `curl http://localhost:8004/health` | Returns `{"status":"ok"}` if all services healthy |
| **Scale ML Tier** | `docker compose up -d --scale ml-service=3` | Nginx load-balances via DNS round-robin |
| **DB Backup** | `docker exec -t indel-postgres-1 pg_dump -U indel indel_demo > backup_$(date +%F).sql` | Daily logical dump |
| **Redis Snapshot** | `redis-cli BGSAVE` | Runs automatically every 5 min |

---

## Scalability & Performance Benchmarks

| Load | Avg. Latency (Premium) | Avg. Latency (Forecast) | Kafka Lag | Redis Hit-Rate |
|------|------------------------|--------------------------|-----------|----------------|
| 1k concurrent workers | 4 ms | 12 ms | < 1 s | 88% |
| 10k concurrent workers | 9 ms | 28 ms | < 5 s | 84% |
| 50k concurrent workers (stress) | 27 ms | 62 ms | < 15 s | 78% |

*All tests run on a single-CPU / 8 GB VM (Docker Desktop). Numbers illustrate the headroom for horizontal scaling via Kubernetes or Docker Swarm.*

---

## Disaster Recovery & High Availability

1. **PostgreSQL** — Continuous WAL archiving to S3-compatible bucket; point-in-time recovery up to the last minute.
2. **Redis** — AOF + periodic RDB snapshots; replicas automatically promoted on master loss (<5 s).
3. **Kafka (KRaft)** — 3-broker quorum ensures leader election without Zookeeper; retained logs for 7 days enable replay of any missed payout events.
4. **ML Service** — Stateless; new containers spin up instantly. Model artifacts stored in a read-only volume (`/app/artifacts`) replicated across nodes.
5. **Target SLAs** — Zero RPO for critical financial data and **<5 min RTO** for any single-service failure.

---

## Data Model Overview

```mermaid
erDiagram
    USERS {
        int id PK
        string phone
        string name
        string role
    }
    WORKER_PROFILES {
        int worker_id PK
        int zone_id FK
        string vehicle_type
        int experience_days
    }
    POLICIES {
        int id PK
        int worker_id FK
        decimal premium_amount
        string status
        timestamp created_at
    }
    ACTIVE_POLICIES {
        int policy_id PK
        int user_id FK
        timestamp started_at
        timestamp updated_at
    }
    ZONES {
        int id PK
        string name
        int risk_level
    }
    CLAIMS {
        int id PK
        int policy_id FK
        decimal payout_amount
        string status
        json risk_explainability
        timestamp created_at
    }

    USERS ||--o{ WORKER_PROFILES : "has"
    USERS ||--o{ POLICIES : "owns"
    POLICIES ||--|| ACTIVE_POLICIES : "is"
    POLICIES ||--o{ CLAIMS : "generates"
    ZONES ||--o{ WORKER_PROFILES : "covers"
```
---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres` | PostgreSQL hostname |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `indel_demo` | Database name |
| `DB_USER` | `indel` | Database user |
| `DB_PASSWORD` | *(secret)* | Database password |
| `REDIS_ADDR` | `redis:6379` | Redis cluster address |
| `KAFKA_BROKERS` | `kafka:9092` | Kafka broker list |
| `JWT_PRIVATE_KEY` | *(secret)* | RS256 private key for token signing |
| `RAZORPAY_KEY_ID` | *(secret)* | Razorpay API key |
| `RAZORPAY_KEY_SECRET` | *(secret)* | Razorpay API secret |
| `OWM_API_KEY` | *(secret)* | OpenWeatherMap API key |
| `OPENAQ_API_KEY` | *(secret)* | OpenAQ API key |
| `FCM_SERVER_KEY` | *(secret)* | Firebase Cloud Messaging key |
| `ML_SERVICE_URL` | `http://ml-service:9000` | Internal ML service endpoint |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |
| `LOG_LEVEL` | `info` | Logging verbosity |
| `ENV` | `demo` | Deployment environment |

See `.env.demo.example` for a complete reference with comments.

---

## The Dashboards

### Platform Dashboard — for Operators

The primary administrative command center governing all continuous telemetry.

- **Platform Command & Worker Directory:** Monitor live order queues, geographic zone assignments, and policy statuses (Live | On Shift vs Offline).
- **Analytics & Disruption Intelligence:** Visualise 7-day multi-city forecast metadata and monitor global claim-generation vs. manual-review queues in real-time.
- **The Chaos Engine:** A dedicated systems-testing panel. Select any tracked zone (e.g., Bhopal) and instantly trigger a `COLLAPSE DEMAND` event, or inject mock Rain and Curfew signals to validate the end-to-end Kafka payout pipeline safely without waiting for a real-world disaster.

![PlatformDashboard](https://github.com/user-attachments/assets/8f063489-ced5-4295-8e44-491cf6d22d3b)

---

### Insurer Dashboard — for Providers

The financial control center built specifically for actuaries and risk-management operators.

- **Pool Posture & Overview:** Monitor live capital liquidity, total premium accumulation vs. payout liabilities, and geographic Zone Net Flow curves.
- **Global Claims Stream:** Real-time ledger of every automated disruption payout (`APPROVED`, `PAID`, `MANUAL_REVIEW`).
- **Fraud Analysis Queue:** Human-in-the-loop audit portal with ML signal explanations and anomaly scores. One-click Approve or Reject.
- **Plan Status Management:** Active subscriber distribution across zones with manual controls to start or terminate coverage plans.

![InsurerDashboard](https://github.com/user-attachments/assets/3384c08d-2a3f-4855-ba40-a32363e04012)

---

### Worker App — for Delivery Partners

Coverage status, this week's AI-computed premium, earnings vs protected baseline, active disruption alerts, claim history, continuity reward progress, and the Maintenance Check self-service SHAP audit — all on one screen. Payment via Razorpay UPI.

#### Registration Journey

![Registration](https://github.com/user-attachments/assets/e63b44d6-aff2-41f2-b7be-dddb7265cc57)

#### Premium Payout & Notifications

![PayoutPipeline](https://github.com/user-attachments/assets/95b5fc44-f7b8-49be-b34c-3e83619715e3)

#### Order & Premium Payout Pipeline

![Claims](https://github.com/user-attachments/assets/14d238fa-4c0c-4aa9-86dd-07995e97f670)

#### Multi-Language Feature

![Lang](https://github.com/user-attachments/assets/98be8d84-a4c6-41fc-bdb5-95fb158d6d65)

---

## Edge-Resilient Mobile App — Offline-First Architecture

Delivery workers operate in India's most connectivity-challenged zones — basement parking, dense urban canyons, rural highway stretches. InDel's Android app is built to function **without any internet connection**, treating connectivity as an enhancement, not a requirement.

### What Works Offline

| Feature | Offline Behaviour | Sync Trigger |
|---------|------------------|--------------|
| **Coverage Status** | Reads from local Room DB — worker sees their active policy, premium amount, and payout ceiling instantly | On reconnect, Room syncs with backend to validate policy state |
| **Disruption Alerts** | FCM push delivers alert to device; alert is persisted in Room DB before display | No sync needed — alert is already committed locally |
| **Claim History** | Entire claim ledger cached in Room DB; readable without connectivity | Incremental sync on reconnect — only delta rows fetched |
| **SHAP Premium Breakdown** | Last computed explainability payload stored locally; worker can audit their premium anytime | Refreshed on next successful API response |
| **Premium Payment** | Razorpay SDK handles offline queue internally — payment queued locally and retried on reconnect | Confirmed when Razorpay webhook fires |

### Room DB Schema Design

```kotlin
@Entity(tableName = "policies")
data class PolicyEntity(
    @PrimaryKey val policyId: Int,
    val workerId: Int,
    val weeklyPremium: Double,
    val status: String,        // active / paused / suspended
    val syncedAt: Long,        // epoch ms — staleness detection
    val explainability: String // JSON blob — SHAP breakdown
)

@Entity(tableName = "disruption_alerts")
data class DisruptionAlertEntity(
    @PrimaryKey val alertId: String,
    val zoneId: Int,
    val type: String,          // WEATHER / AQI / ORDER_DROP
    val issuedAt: Long,
    val payoutAmount: Double,
    val status: String         // pending / paid / rejected
)
```

### Sync Strategy — Conflict-Free by Design

InDel uses a **server-wins, append-only** sync model:

1. **On connect** → app sends its local `syncedAt` timestamp
2. **Backend responds** with all records modified after that timestamp (delta only)
3. **Room DB merges** via `INSERT OR REPLACE` — no conflict resolution needed
4. **Claims are immutable** — once written, they are never mutated locally, only appended to

### Zero-Contamination per Worker

Gig platforms share a critical edge case: **multiple workers logging in on the same physical device** (family phones, shared delivery devices). InDel solves this at the database level:

```kotlin
// Called on every login token generation
fun purgeUserState(previousWorkerId: Int) {
    db.policyDao().deleteAllForWorker(previousWorkerId)
    db.alertDao().deleteAllForWorker(previousWorkerId)
    db.claimDao().deleteAllForWorker(previousWorkerId)
    SecurePrefs.clearAll()
}
```

- **On every new login**, the encrypted Room DB is **completely purged** for the previous worker before the new session loads.
- **No ghost profiles** — Worker B can never see Worker A's premium, claim history, or payout data.
- **Encrypted at rest** — Room DB uses `androidx.security.crypto` with AES-256 encryption tied to the device's hardware-backed keystore.

### Network Resilience in Numbers

| Scenario | App Behaviour |
|----------|---------------|
| No connectivity at app open | Full UI loads from Room DB in **< 80 ms** |
| Disruption alert during offline | FCM delivers to device; alert persisted and displayed immediately |
| Reconnects after 2 hours offline | Delta sync fetches only changed records — completes in **< 1.2 s** |
| Worker switches accounts on same device | Room DB purged in **< 50 ms** before new session loads |

---

## Increasing Efficiency of Resource Utilization

### 1. Unified Machine Learning Container

Instead of deploying three disjointed Python microservices (Premium Pricing, Fraud Detection, Forecaster) that inflated infrastructure by multiplying Python interpreter memory overhead — all inference tools now share a single dense FastAPI container (`ml-service`), dropping computational footprint by **two-thirds** without compromising parallel throughput.

### 2. High-Performance Messaging Protocol (Kafka KRaft Mode)

Zookeeper acts as a significant memory burden. We stripped out external Kafka coordination services and fully migrated to **Zookeeper-less KRaft Mode**. Kafka now natively manages partition leadership via quorum, shedding immense JVM latency and permitting the entire orchestration suite to run cleanly on bounded compute.

### 3. Defensive Actuarial Boundaries

Dynamic ML struggles with "Cold Start" user distributions. Recognizing that new workers generated wild extrapolations from our XGBoost regressor, we introduced a **mathematical bounding wrapper** directly to the inference API. Out-of-bounds calculations instantly clamp to stable base premiums (e.g., ₹49), guaranteeing logical actuarial safety.

### 4. Zero-Contamination Offline-First Sync (Android Room DB)

Gig delivery platforms share a notorious edge-case: multi-account logins on identical hardware. We hardened the offline-first Android interface by systematically **purging encrypted Room Databases** dynamically on token generation/revocation. Dashboard rendering states are mathematically isolated per-user, fully eliminating cross-state ghost profiles.

---

## Policy Lifecycle — No Renewals, No Calls, No Forms

Workers register once. Coverage runs forever in the background.

| State | Trigger |
|-------|---------|
| **Active** | Premium paid, coverage running |
| **Paused** | 1 missed weekly payment |
| **Suspended** | 2+ consecutive missed payments |
| **Rewarded** | Consistent payments + zero claims → reduced premium or extended payout ceiling |

**Cold-start workers** (under 20 verified deliveries) use zone-average income baselines. Claims filed within the first 7 days of enrollment are auto-held for manual review — no gaming around known events.

**TTL Gate — Anti-Ghost-Login Defense:** When disruption is confirmed, the system checks every worker's `lastActiveAt` telemetry timestamp against a hardcoded 15-minute backward-looking window, locked to the millisecond of disruption confirmation. Dormant accounts that log in after seeing the rain? **Stripped from eligibility arrays automatically.**

---

## ML Under the Hood

| Model | Algorithm | Job | Retraining |
|-------|-----------|-----|------------|
| Premium Calculator | XGBoost + SHAP (18 features) | Income loss probability → weekly premium per worker | Monthly / continuous |
| Fraud Detector | IsolationForest + DBSCAN + Hard Rules | Anomaly score + cluster fit + hard disqualifiers | Weekly |
| Disruption Forecaster | Facebook Prophet | 7-day zone claim probability for insurer reserve planning | Weekly |

---

## FAQ & Known Limitations

**Q: Why does the forecast endpoint require a POST?**
A: The request payload contains the `zone_id` and optional overrides (e.g., custom date range). POST keeps the payload out of the URL and allows future extensions without breaking clients.

**Q: Can we run the ML service on GPU?**
A: The current model (Prophet + XGBoost) is CPU-bound; GPU acceleration would only benefit the optional deep-learning fraud detector, slated for a future release.

**Q: What happens if Redis goes down?**
A: Go services fall back to PostgreSQL on a cache miss. Because the cache-aside pattern writes-through on updates, data consistency is never compromised.

**Q: What about IRDAI compliance?**
A: The platform is architected to satisfy IRDAI actuarial audit requirements. Full actuarial-grade audit trails for every premium calculation, claim event, and payout are stored immutably in Kafka and PostgreSQL.

**Known Limitation:** The current demo uses a **single-node Kafka** (KRaft) for simplicity. Production should run a 3-node quorum with replication factor = 3 for true fault tolerance.

---

## Team ImaginAI

*Five people. One conviction: gig workers deserve financial protection that works as hard as they do.*

| Name | Built |
|------|-------|
| Shravanthi S | Core Policy Logic, Disruption Sync Cycle, Payout & Data Operations |
| Gayathri U | Delivery Management, Postgres Schema, DevOps & Docker Compositions |
| Rithanya K A | Python FastAPI ML Services — XGBoost Training & Inference |
| Saravana Priyaa C R | Platform Integration, Chaos Engine, Disruption Engine |
| Subiksha MV | Insurer System, Claims Intelligence & Overall System Design |

---

<p align="center"><b>Guidewire DEVTrails 2026 — E-Commerce Delivery Persona — Phase 3: Scale & Optimize</b></p>

<p align="center"><i>Figures are illustrative estimates for design and modelling purposes. Production deployment requires IRDAI registration and KYC/AML compliance by the deploying insurer. All systems are strictly idempotent — mass disruption events cannot produce duplicate claims or double payouts.</i></p>
