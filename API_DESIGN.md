# InDel — API Design

> Version: v1
> All endpoints prefixed with `/api/v1/`
> All requests and responses in JSON unless stated otherwise
> All authenticated endpoints require `Authorization: Bearer <jwt_token>` header
> All timestamps are ISO 8601 in UTC

---

## Global Conventions

### ID Format

All entity IDs follow a prefixed UUID format:

| Entity | Format | Example |
|---|---|---|
| Worker | `wkr_<uuid>` | `wkr_9f8e7d6c` |
| Claim | `clm_<uuid>` | `clm_def456` |
| Disruption | `dis_<uuid>` | `dis_xyz789` |
| Policy | `pol_<uuid>` | `pol_abc123` |
| Payout | `pay_<uuid>` | `pay_ghi012` |
| Order | `ord_<uuid>` | `ord_swg_abc123` |
| Zone | `zone_<slug>` | `tambaram_chennai` |

---

### Pagination Standard

All paginated endpoints return this wrapper:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 124,
    "has_next": true
  }
}
```

Query params for paginated endpoints: `?page=1&limit=20`

---

### Enum Definitions

**Claim Status:**
- `pending` — generated, awaiting fraud scoring
- `approved` — fraud check passed, payout queued
- `rejected` — failed eligibility or hard rule violation
- `fraud_flagged` — routed to manual review queue

**Fraud Verdict:**
- `clean` — all three layers passed
- `review` — medium risk, delayed validation
- `fraud` — high risk, flagged or auto-rejected

**Disruption Type:**
- `heavy_rain`
- `extreme_heat`
- `severe_aqi`
- `curfew`
- `bandh`
- `zone_closure`
- `order_drop`
- `flash_flood`

**Disruption Severity:**
- `low`
- `medium`
- `high`
- `critical`

**Policy Status:**
- `active`
- `paused`
- `suspended`
- `cancelled`

**Payout Status:**
- `queued`
- `processing`
- `credited`
- `failed`
- `retrying`

**Vehicle Type:**
- `two_wheeler`
- `three_wheeler`
- `four_wheeler`
- `bicycle`

**Baseline Source:**
- `worker_history` — 4-week personal earnings history
- `zone_average` — cold start, derived from zone peer group
- `peer_group_average` — cold start, same zone and vehicle type

---

### Timestamp Standard

All timestamps are ISO 8601 in UTC:

```
2026-03-21T11:40:00Z
```

Never use local time. Never use Unix epoch in responses. Unix epoch is acceptable internally in Kafka message headers only.

---

### Idempotency

All payout and premium payment endpoints require an `Idempotency-Key` header:

```http
Idempotency-Key: pay_clm_def456
```

- Key must be unique per operation
- Same key submitted twice returns the original response without re-processing
- Keys expire after 24 hours
- Missing key on a payment endpoint returns `400 IDEMPOTENCY_KEY_REQUIRED`

---

### Gateway Role

Each gateway enforces role-based access control and isolates client-specific logic. The Worker Gateway never exposes insurer economics. The Insurer Gateway never exposes raw worker PII beyond what is necessary for claim review. The Platform Gateway is scoped to delivery operations only.

---

## Architecture Overview

```
Kotlin Worker App          →  Worker Gateway        (Go) → :8001
Insurer Dashboard          →  Insurer Gateway       (Go) → :8002
Delivery Platform          →  Platform Gateway      (Go) → :8003

All Gateways               →  Core Backend Service  (Go) → :8000
Core Backend               →  ML Microservices      (Python FastAPI)
                               ├── Premium Service         → :9001
                               ├── Fraud Service           → :9002
                               └── Forecast Service        → :9003

Core Backend               →  PostgreSQL            → :5432
Core Backend               →  Kafka                 → :9092
Core Backend               →  Razorpay Sandbox
Core Backend               →  OpenWeatherMap API
Core Backend               →  OpenAQ API
```

---

## Health & Status Endpoints

> Available on every service. No auth required. Used for CI/CD checks, demo verification, and debugging.

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/health` | Liveness check — is the service running? |
| GET | `/api/v1/status` | Dependency status — are all connections healthy? |

**GET /api/v1/health — Response:**
```json
{
  "service": "worker-gateway",
  "status": "healthy",
  "timestamp": "2026-03-21T11:40:00Z"
}
```

**GET /api/v1/status — Response:**
```json
{
  "service": "core-backend",
  "status": "healthy",
  "dependencies": {
    "postgres": "connected",
    "kafka": "connected",
    "premium_service": "reachable",
    "fraud_service": "reachable",
    "forecast_service": "reachable",
    "openweathermap": "reachable",
    "openaq": "reachable",
    "razorpay": "reachable"
  },
  "timestamp": "2026-03-21T11:40:00Z"
}
```

> If any dependency is `unreachable`, the overall status returns `degraded` not `unhealthy` — the system continues operating with fallback logic (internal InDel signals replace external API signals per the README design).

---

## Demo Endpoints

> Available in non-production environments only. Removed from production build via environment flag `INDEL_ENV=demo`.
> Used to simulate disruption events during hackathon demo without waiting for real weather API thresholds.

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/demo/trigger-disruption` | Simulate a disruption event in a zone |
| POST | `/api/v1/demo/settle-earnings` | Simulate weekly earnings settlement for a worker |
| POST | `/api/v1/demo/reset-zone` | Reset zone to pre-disruption state |

**POST /api/v1/demo/trigger-disruption — Request:**
```json
{
  "zone": "tambaram_chennai",
  "type": "heavy_rain",
  "severity": "high",
  "confidence_score": 0.91,
  "duration_hours": 5.83
}
```

**POST /api/v1/demo/trigger-disruption — Response:**
```json
{
  "disruption_id": "dis_xyz789",
  "zone": "tambaram_chennai",
  "status": "confirmed",
  "eligible_workers": 12,
  "claims_generated": 12,
  "message": "Disruption confirmed. Claims generating for eligible workers."
}
```

> These endpoints give full control over the demo flow. If the OpenWeatherMap API is slow or unresponsive during the live demo, the demo trigger bypasses it entirely and fires the full internal pipeline.

---

## Authentication

### JWT Token Structure

```json
{
  "sub": "worker_id | insurer_id | platform_id",
  "role": "worker | insurer_admin | platform_admin | super_admin",
  "zone": "tambaram_chennai",
  "iat": 1711234567,
  "exp": 1711320967
}
```

### Auth Endpoints (shared across all gateways)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, returns JWT |
| POST | `/api/v1/auth/refresh` | Refresh JWT token |
| POST | `/api/v1/auth/logout` | Invalidate token |
| POST | `/api/v1/auth/otp/send` | Send Firebase OTP to phone |
| POST | `/api/v1/auth/otp/verify` | Verify OTP, returns JWT |

---

## Gateway 1 — Worker API (:8001)

> Consumed by: Kotlin Android App
> Auth: Worker JWT

---

### Onboarding

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/worker/onboard` | Register worker profile |
| GET | `/api/v1/worker/profile` | Get worker profile |
| PUT | `/api/v1/worker/profile` | Update worker profile |
| POST | `/api/v1/worker/kyc` | Submit KYC details (mocked) |

**POST /api/v1/worker/onboard — Request:**
```json
{
  "name": "Priya Rajan",
  "phone": "+919876543210",
  "home_zone": "tambaram_chennai",
  "vehicle_type": "two_wheeler",
  "upi_id": "priya@upi",
  "device_id": "a1b2c3d4e5f6",
  "preferred_hours": {
    "start": "09:00",
    "end": "21:00"
  }
}
```

**POST /api/v1/worker/onboard — Response:**
```json
{
  "worker_id": "wkr_9f8e7d6c",
  "status": "onboarded",
  "coverage_eligible_from": "2026-04-01T00:00:00Z"
}
```

---

### Policy & Coverage

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/worker/policy/enroll` | Opt into income protection |
| GET | `/api/v1/worker/policy` | Get active policy details |
| PUT | `/api/v1/worker/policy/pause` | Pause coverage |
| PUT | `/api/v1/worker/policy/cancel` | Cancel policy |
| GET | `/api/v1/worker/policy/premium` | Get current week premium |
| POST | `/api/v1/worker/policy/premium/pay` | Manual premium payment |

**POST /api/v1/worker/policy/premium/pay — Headers:**
```http
Idempotency-Key: pay_pol_abc123_week_2026-03-17
```

**GET /api/v1/worker/policy — Response:**
```json
{
  "policy_id": "pol_abc123",
  "status": "active",
  "coverage_start": "2026-03-17T00:00:00Z",
  "current_week": {
    "start": "2026-03-17",
    "end": "2026-03-23",
    "premium": 22.00,
    "premium_paid": true,
    "max_payout": 800.00,
    "coverage_ratio": 0.85
  },
  "continuity_streak_weeks": 4,
  "next_reward_at_weeks": 8
}
```

---

### Earnings

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/worker/earnings` | Get earnings summary |
| GET | `/api/v1/worker/earnings/history` | Earnings history paginated |
| GET | `/api/v1/worker/earnings/baseline` | Get 4-week baseline rate |

**GET /api/v1/worker/earnings — Response:**
```json
{
  "this_week": {
    "actual": 2400.00,
    "baseline": 4200.00,
    "protected_income": 3570.00,
    "shortfall": 1800.00
  },
  "baseline_hourly_rate": 120.00,
  "active_hours_today": 6.5
}
```

---

### Disruption & Claims

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/worker/disruptions/active` | Get active disruptions in worker zone |
| GET | `/api/v1/worker/claims` | Get claim history paginated |
| GET | `/api/v1/worker/claims/:claim_id` | Get single claim detail |
| POST | `/api/v1/worker/claims/:claim_id/maintenance` | Trigger maintenance check |
| GET | `/api/v1/worker/claims/:claim_id/maintenance/:check_id` | Get maintenance check result |

**GET /api/v1/worker/disruptions/active — Response:**
```json
{
  "disruptions": [
    {
      "disruption_id": "dis_xyz789",
      "zone": "tambaram_chennai",
      "type": "heavy_rain",
      "severity": "high",
      "started_at": "2026-03-21T11:40:00Z",
      "estimated_end": "2026-03-21T17:30:00Z",
      "confidence_score": 0.91,
      "triggers": ["weather_alert", "order_drop_detected"],
      "worker_eligible": true
    }
  ]
}
```

**GET /api/v1/worker/claims/:claim_id — Response:**
```json
{
  "claim_id": "clm_def456",
  "disruption_id": "dis_xyz789",
  "status": "approved",
  "disruption_type": "heavy_rain",
  "disruption_window": {
    "start": "2026-03-21T11:40:00Z",
    "end": "2026-03-21T17:30:00Z",
    "hours": 5.83
  },
  "income_loss": {
    "expected": 700.00,
    "actual": 80.00,
    "loss": 620.00,
    "coverage_ratio": 0.85,
    "payout": 527.00
  },
  "fraud_score": 0.12,
  "fraud_verdict": "clean",
  "payout_status": "credited",
  "payout_method": "upi",
  "credited_at": "2026-03-21T18:05:00Z"
}
```

---

### Payouts & Wallet

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/worker/wallet` | Get wallet balance |
| GET | `/api/v1/worker/payouts` | Get payout history paginated |
| POST | `/api/v1/worker/payouts/:payout_id/confirm` | Confirm payout (optional worker confirmation) |

---

### Notifications & Preferences

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/worker/notifications` | Get notifications paginated |
| PUT | `/api/v1/worker/preferences` | Update language and notification preferences |
| POST | `/api/v1/worker/device/register` | Register FCM push token |

---

## Gateway 2 — Insurer API (:8002)

> Consumed by: Vite + React + Tremor Dashboard
> Auth: Insurer Admin JWT

---

### Overview & KPIs

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/overview` | Top-level KPI summary |
| GET | `/api/v1/insurer/pool/health` | Premium pool health this week |

**GET /api/v1/insurer/overview — Response:**
```json
{
  "active_workers": 1000,
  "active_policies": 847,
  "this_week": {
    "premiums_collected": 68000.00,
    "payouts_disbursed": 44000.00,
    "gross_margin": 24000.00,
    "loss_ratio": 0.647
  },
  "pending_claims": 12,
  "fraud_flagged_claims": 3,
  "reserve_balance": 210000.00
}
```

---

### Loss Ratio

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/loss-ratio` | Loss ratio by zone and city |
| GET | `/api/v1/insurer/loss-ratio/history` | Historical loss ratio trend paginated |

**GET /api/v1/insurer/loss-ratio — Response:**
```json
{
  "overall": 0.647,
  "by_city": [
    { "city": "Chennai", "loss_ratio": 0.72, "risk_profile": "high" },
    { "city": "Bengaluru", "loss_ratio": 0.61, "risk_profile": "medium" },
    { "city": "Pune", "loss_ratio": 0.54, "risk_profile": "low" }
  ],
  "by_zone": [
    { "zone": "tambaram_chennai", "loss_ratio": 0.74, "active_policies": 124 },
    { "zone": "koramangala_bengaluru", "loss_ratio": 0.58, "active_policies": 89 }
  ]
}
```

---

### Claims Pipeline

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/claims` | All claims paginated with filters |
| GET | `/api/v1/insurer/claims/:claim_id` | Single claim full detail |
| PUT | `/api/v1/insurer/claims/:claim_id/approve` | Manual approve claim |
| PUT | `/api/v1/insurer/claims/:claim_id/reject` | Manual reject claim |
| GET | `/api/v1/insurer/claims/fraud-queue` | Fraud-flagged claims queue paginated |
| PUT | `/api/v1/insurer/claims/:claim_id/fraud/resolve` | Resolve fraud-flagged claim |

**GET /api/v1/insurer/claims — Query params:**
```
?status=pending|approved|rejected|fraud_flagged
&zone=tambaram_chennai
&disruption_type=heavy_rain|extreme_heat|severe_aqi|curfew|order_drop
&from=2026-03-01
&to=2026-03-21
&page=1
&limit=20
```

**GET /api/v1/insurer/claims/fraud-queue — Response:**
```json
{
  "data": [
    {
      "claim_id": "clm_fraud01",
      "worker_id": "wkr_abc123",
      "disruption_type": "heavy_rain",
      "payout_amount": 527.00,
      "fraud_score": 0.87,
      "fraud_signals": [
        "gps_zone_mismatch",
        "claim_frequency_high",
        "deviation_from_zone_cluster"
      ],
      "isolation_forest_score": 0.89,
      "dbscan_verdict": "noise_point",
      "rule_violations": [],
      "flagged_at": "2026-03-21T14:22:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 3,
    "has_next": false
  }
}
```

---

### Forecast & Reserve

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/forecast` | 7-day claim probability by zone |
| GET | `/api/v1/insurer/reserve` | Reserve recommendation |

**GET /api/v1/insurer/forecast — Response:**
```json
{
  "generated_at": "2026-03-21T06:00:00Z",
  "forecast_window": {
    "from": "2026-03-22",
    "to": "2026-03-28"
  },
  "by_zone": [
    {
      "zone": "tambaram_chennai",
      "claim_probability": 0.74,
      "expected_claims": 18,
      "expected_payout": 9900.00,
      "risk_drivers": ["monsoon_proximity", "historical_flood_frequency"]
    }
  ],
  "aggregate": {
    "expected_claims": 56,
    "expected_payout": 30800.00,
    "recommended_reserve": 46200.00
  },
  "model_note": "Per-zone forecast only. Cross-zone correlation not modelled in prototype. Aggregate figures may underestimate correlated mass disruption events."
}
```

---

### Workers & Policies

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/workers` | All workers paginated |
| GET | `/api/v1/insurer/workers/:worker_id` | Single worker detail |
| GET | `/api/v1/insurer/policies` | All active policies paginated |
| GET | `/api/v1/insurer/policies/:policy_id` | Single policy detail |

---

### Maintenance Check Queue

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/maintenance-checks` | All pending maintenance checks paginated |
| GET | `/api/v1/insurer/maintenance-checks/:check_id` | Single check with AI output and worker data |
| POST | `/api/v1/insurer/maintenance-checks/:check_id/respond` | Send reviewer response to worker |

---

### Premium & Pricing

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/insurer/premiums/summary` | Premium collection summary |
| GET | `/api/v1/insurer/premiums/zone-rates` | Current premium rates by zone |
| PUT | `/api/v1/insurer/premiums/zone-rates/:zone` | Override zone premium rate |

---

## Gateway 3 — Delivery Platform API (:8003)

> Consumed by: Delivery Platform Dashboard + Swiggy/Zomato webhook integration
> Auth: Platform Admin JWT + Webhook HMAC signature

---

### Worker Management

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/platform/workers` | All InDel-enrolled workers paginated |
| GET | `/api/v1/platform/workers/:worker_id` | Single worker delivery profile |
| GET | `/api/v1/platform/workers/:worker_id/coverage` | Worker coverage status |
| GET | `/api/v1/platform/zones` | All active zones with worker counts |

---

### Order Integration (Webhooks from Platform)

> These endpoints receive webhooks from Swiggy/Zomato or simulate them for demo purposes.
> All webhook requests must include `X-InDel-Signature` HMAC header.

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/platform/webhooks/order/assigned` | Order assigned to worker |
| POST | `/api/v1/platform/webhooks/order/completed` | Order completed by worker |
| POST | `/api/v1/platform/webhooks/order/cancelled` | Order cancelled |
| POST | `/api/v1/platform/webhooks/earnings/settled` | Weekly earnings settled — triggers premium prompt |

**POST /api/v1/platform/webhooks/order/completed — Request:**
```json
{
  "order_id": "ord_swg_abc123",
  "worker_id": "wkr_9f8e7d6c",
  "zone": "tambaram_chennai",
  "completed_at": "2026-03-21T13:45:00Z",
  "earnings": 85.00,
  "platform": "swiggy"
}
```

**POST /api/v1/platform/webhooks/earnings/settled — Request:**
```json
{
  "worker_id": "wkr_9f8e7d6c",
  "week_start": "2026-03-17",
  "week_end": "2026-03-23",
  "total_earnings": 4200.00,
  "total_orders": 49,
  "settled_at": "2026-03-24T09:00:00Z"
}
```

> This webhook triggers the premium deduction prompt in the worker app — the earnings settlement moment is when the premium prompt appears, not before. This is the payment timing design: the worker receives their earnings and sees the premium prompt at the highest-trust, lowest-resistance moment in their financial cycle.

---

### Zone & Disruption Monitoring

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/platform/disruptions/active` | Active disruptions across all zones |
| GET | `/api/v1/platform/zones/:zone_id/orders` | Live order volume for zone |
| GET | `/api/v1/platform/zones/:zone_id/workers/active` | Active workers in zone right now |

---

### Dashboard Analytics

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/platform/analytics/overview` | Platform-level delivery summary |
| GET | `/api/v1/platform/analytics/zones` | Zone-level performance metrics paginated |

---

## Internal — Core Backend Service (:8000)

> Not exposed externally. Called only by the three gateways.
> Handles all business logic, database writes, Kafka publishing, and ML service calls.

---

### Disruption Engine (Internal)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/internal/v1/disruptions/evaluate` | Evaluate zone disruption candidate |
| POST | `/internal/v1/disruptions/confirm` | Confirm disruption, open window |
| POST | `/internal/v1/disruptions/close` | Close disruption window |
| GET | `/internal/v1/disruptions/:disruption_id/eligible-workers` | Get eligible workers for disruption |

---

### Claim Engine (Internal)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/internal/v1/claims/generate` | Auto-generate claim for eligible worker |
| POST | `/internal/v1/claims/:claim_id/score` | Run fraud scoring pipeline |
| POST | `/internal/v1/claims/:claim_id/route` | Route claim based on fraud score |
| POST | `/internal/v1/claims/:claim_id/payout` | Queue payout via Kafka |

**POST /internal/v1/claims/:claim_id/payout — Headers:**
```http
Idempotency-Key: pay_clm_def456
```

---

### Income Loss Engine (Internal)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/internal/v1/income/baseline` | Calculate worker earnings baseline |
| POST | `/internal/v1/income/loss` | Calculate income loss for disruption window |

**POST /internal/v1/income/loss — Request:**
```json
{
  "worker_id": "wkr_9f8e7d6c",
  "disruption_id": "dis_xyz789",
  "disruption_window": {
    "start": "2026-03-21T11:40:00Z",
    "end": "2026-03-21T17:30:00Z"
  }
}
```

**POST /internal/v1/income/loss — Response:**
```json
{
  "baseline_hourly_rate": 120.00,
  "disruption_hours": 5.83,
  "expected_earnings": 700.00,
  "actual_earnings": 80.00,
  "income_loss": 620.00,
  "coverage_ratio": 0.85,
  "payout_amount": 527.00,
  "cold_start": false,
  "baseline_source": "worker_history"
}
```

---

## Internal — ML Microservices

---

### Premium Service (:9001)

> Called by Core Backend when a worker enrolls or at weekly premium recalculation.

**POST /ml/v1/premium/calculate**

Request:
```json
{
  "worker_id": "wkr_9f8e7d6c",
  "zone": "tambaram_chennai",
  "features": {
    "zone_disruption_frequency_24m": 0.34,
    "seasonal_risk_score": 0.71,
    "rolling_aqi_4w": 142.0,
    "worker_daily_active_hours": 8.5,
    "order_density_variance": 0.22,
    "income_stability_score": 0.68
  }
}
```

Response:
```json
{
  "worker_id": "wkr_9f8e7d6c",
  "weekly_premium": 22.00,
  "risk_score": 0.71,
  "shap_breakdown": {
    "flood_risk_zone": 6.00,
    "rolling_aqi_pattern": 3.00,
    "income_instability": 2.00,
    "base_rate": 7.00,
    "seasonal_adjustment": 4.00
  },
  "plain_language": "Your premium is ₹22 mainly because your area has a high flood risk (₹6) and it is monsoon season (₹4)."
}
```

---

### Fraud Service (:9002)

> Called by Core Backend claim scoring pipeline.

**POST /ml/v1/fraud/score**

Request:
```json
{
  "claim_id": "clm_def456",
  "worker_id": "wkr_9f8e7d6c",
  "disruption_id": "dis_xyz789",
  "features": {
    "gps_in_zone": true,
    "gps_zone_match_confidence": 0.94,
    "completed_orders_during_disruption": 0,
    "claimed_loss_to_baseline_ratio": 0.89,
    "claim_frequency_8w": 1,
    "zone_claim_clustering_score": 0.87,
    "mobility_pattern_score": 0.91,
    "days_since_enrollment": 45
  }
}
```

Response:
```json
{
  "claim_id": "clm_def456",
  "isolation_forest_score": 0.12,
  "dbscan_verdict": "in_cluster",
  "rule_violations": [],
  "overall_fraud_score": 0.12,
  "verdict": "clean",
  "routing": "auto_approve",
  "explanation": "Claim pattern consistent with zone-wide disruption impact."
}
```

---

### Forecast Service (:9003)

> Called by Core Backend on weekly schedule for insurer dashboard.

**POST /ml/v1/forecast/zone**

Request:
```json
{
  "zone": "tambaram_chennai",
  "forecast_days": 7,
  "features": {
    "historical_weather": [],
    "historical_aqi": [],
    "historical_order_volume": [],
    "season": "pre_monsoon"
  }
}
```

Response:
```json
{
  "zone": "tambaram_chennai",
  "forecast": [
    {
      "date": "2026-03-22",
      "claim_probability": 0.31,
      "expected_claims": 4,
      "expected_payout": 2200.00
    },
    {
      "date": "2026-03-23",
      "claim_probability": 0.74,
      "expected_claims": 18,
      "expected_payout": 9900.00
    }
  ],
  "week_aggregate": {
    "expected_claims": 56,
    "expected_payout": 30800.00,
    "recommended_reserve": 46200.00
  },
  "model": "prophet",
  "limitation": "Per-zone forecast only. Cross-zone correlation not modelled in prototype."
}
```

---

## Kafka Topics

| Topic | Producer | Consumer | Purpose |
|---|---|---|---|
| `indel.weather.alerts` | OpenWeatherMap poller | Core Backend | Raw weather event ingestion |
| `indel.aqi.alerts` | OpenAQ poller | Core Backend | AQI event ingestion |
| `indel.zone.order-drop` | Platform webhook handler | Core Backend | Zone order volume anomaly |
| `indel.disruption.confirmed` | Core Backend | Claim Engine | Disruption window opened |
| `indel.claims.generated` | Claim Engine | Fraud Service | New claim ready for scoring |
| `indel.claims.scored` | Fraud Service | Claim Router | Fraud score ready |
| `indel.payouts.queued` | Claim Router | Payout Processor | Approved claim ready for payout |
| `indel.payouts.completed` | Payout Processor | Notification Service | Payout credited, notify worker |
| `indel.payouts.failed` | Payout Processor | Payout Processor | Retry failed payout — exponential backoff |

---

## Standard Error Format

All error responses follow this structure:

```json
{
  "error": {
    "code": "WORKER_NOT_ELIGIBLE",
    "message": "Worker was not active in zone before disruption window opened.",
    "details": {
      "worker_id": "wkr_9f8e7d6c",
      "disruption_id": "dis_xyz789",
      "reason": "no_activity_before_trigger"
    }
  },
  "request_id": "req_abc123",
  "timestamp": "2026-03-21T14:22:00Z"
}
```

### Error Codes

| Code | HTTP Status | Meaning |
|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `FORBIDDEN` | 403 | Valid JWT but wrong role |
| `IDEMPOTENCY_KEY_REQUIRED` | 400 | Payment endpoint called without idempotency key |
| `WORKER_NOT_FOUND` | 404 | Worker ID does not exist |
| `POLICY_NOT_ACTIVE` | 400 | Worker has no active policy |
| `WORKER_NOT_ELIGIBLE` | 400 | Worker does not meet eligibility criteria |
| `DISRUPTION_NOT_CONFIRMED` | 400 | Disruption confidence below threshold |
| `CLAIM_ALREADY_EXISTS` | 409 | Duplicate claim for same disruption |
| `FRAUD_FLAGGED` | 400 | Claim routed to manual review |
| `PREMIUM_UNPAID` | 402 | Coverage inactive due to missed premium |
| `COLD_START_HOLD` | 400 | Claim held — worker enrolled less than 7 days ago |
| `PAYOUT_FAILED` | 500 | Payment gateway error — queued for retry |
| `ORACLE_CONSENSUS_FAILED` | 503 | Insufficient external signal confidence |

---

## API Versioning Policy

- Current version: `v1`
- All endpoints prefixed `/api/v1/`
- Breaking changes increment to `v2` — `v1` remains supported for 6 months
- Non-breaking additions (new fields, new endpoints) do not increment version
- Deprecation notices added to response headers: `X-InDel-Deprecated: true`

---

## Rate Limits

| Gateway | Limit | Window |
|---|---|---|
| Worker API | 100 requests | per minute per worker |
| Insurer API | 300 requests | per minute per insurer |
| Platform API | 500 requests | per minute per platform |
| Internal APIs | No limit | Internal network only |
| ML Microservices | No limit | Internal network only |

---

*API Design v1 — Team ImaginAI — Guidewire DEVTrails 2026*
*Note: Request/response schemas are illustrative. Final field names and types will be confirmed during implementation.*