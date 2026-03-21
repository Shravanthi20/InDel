# InDel — Database Design

> Database: PostgreSQL
> ORM: GORM (Go backend), SQLAlchemy (Python ML services)
> Migrations: golang-migrate
> All tables use UUID primary keys matching the ID format conventions in API_DESIGN.md
> All timestamps are UTC

---

## Table of Contents

1. [Schema Overview](#schema-overview)
2. [Users & Identity](#users--identity)
3. [Zones](#zones)
4. [Workers](#workers)
5. [Policies & Premiums](#policies--premiums)
6. [Orders & Earnings](#orders--earnings)
7. [Disruptions](#disruptions)
8. [Claims & Payouts](#claims--payouts)
9. [ML Model Outputs](#ml-model-outputs)
10. [Notifications](#notifications)
11. [Audit & System](#audit--system)
12. [Indexes](#indexes)
13. [Entity Relationship Summary](#entity-relationship-summary)

---

## Schema Overview

```
workers ──────────────────────────────────────────────────────┐
   │                                                           │
   ├──── policies ──── premium_payments                        │
   │                                                           │
   ├──── orders ──── earnings_records                          │
   │         │                                                 │
   │         └──── weekly_earnings_summaries                   │
   │                                                           │
   ├──── disruption_eligibility ──── disruptions ──── zones    │
   │                                                           │
   ├──── claims ──── claim_fraud_scores                        │
   │         │                                                 │
   │         ├──── payouts                                     │
   │         └──── maintenance_checks                          │
   │                                                           │
   ├──── premium_model_outputs                                 │
   ├──── fraud_model_outputs                                   │
   └──── notifications                                         │
                                                               │
insurer_admins                                                 │
platform_admins                                                │
zones ─────────────────────────────────────────────────────────┘
```

---

## Users & Identity

### Table: `worker_users`

> One record per registered delivery worker.

```sql
CREATE TABLE worker_users (
    id              VARCHAR(20) PRIMARY KEY,          -- wkr_<uuid>
    phone           VARCHAR(15) NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    device_id       VARCHAR(100) NOT NULL UNIQUE,     -- fraud: 1 device per worker
    upi_id          VARCHAR(100),
    bank_account    VARCHAR(50),
    kyc_status      VARCHAR(20) NOT NULL DEFAULT 'pending',
                    -- pending | submitted | verified | rejected
    kyc_aadhaar     VARCHAR(12),                      -- mocked in prototype
    kyc_pan         VARCHAR(10),                      -- mocked in prototype
    language_pref   VARCHAR(10) NOT NULL DEFAULT 'en',
    fcm_token       VARCHAR(255),                     -- Firebase push token
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP                                -- soft delete
);
```

---

### Table: `insurer_admin_users`

```sql
CREATE TABLE insurer_admin_users (
    id              VARCHAR(20) PRIMARY KEY,          -- ins_<uuid>
    email           VARCHAR(100) NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    insurer_name    VARCHAR(100) NOT NULL,
    role            VARCHAR(30) NOT NULL DEFAULT 'insurer_admin',
                    -- insurer_admin | insurer_super_admin
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `platform_admin_users`

```sql
CREATE TABLE platform_admin_users (
    id              VARCHAR(20) PRIMARY KEY,          -- plt_<uuid>
    email           VARCHAR(100) NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    platform_name   VARCHAR(50) NOT NULL,             -- swiggy | zomato
    webhook_secret  VARCHAR(100) NOT NULL,            -- HMAC secret for webhook verification
    role            VARCHAR(30) NOT NULL DEFAULT 'platform_admin',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `auth_tokens`

> Tracks active JWT refresh tokens for invalidation on logout.

```sql
CREATE TABLE auth_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(20) NOT NULL,
    user_type       VARCHAR(20) NOT NULL,             -- worker | insurer_admin | platform_admin
    refresh_token   VARCHAR(500) NOT NULL UNIQUE,
    expires_at      TIMESTAMP NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Zones

### Table: `zones`

> Seeded from existing delivery app zone data at setup. Not created dynamically.

```sql
CREATE TABLE zones (
    id              VARCHAR(50) PRIMARY KEY,          -- tambaram_chennai (slug format)
    display_name    VARCHAR(100) NOT NULL,            -- "Tambaram, Chennai"
    city            VARCHAR(50) NOT NULL,             -- Chennai | Bengaluru | Pune
    state           VARCHAR(50) NOT NULL,
    lat_center      DECIMAL(9,6) NOT NULL,
    lng_center      DECIMAL(9,6) NOT NULL,
    risk_profile    VARCHAR(10) NOT NULL DEFAULT 'medium',
                    -- low | medium | high | critical
    base_premium    DECIMAL(8,2) NOT NULL,            -- base before ML adjustment
    max_payout      DECIMAL(8,2) NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `zone_risk_history`

> Zone risk score updated after every disruption event. Used by Prophet forecaster.

```sql
CREATE TABLE zone_risk_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    risk_score      DECIMAL(4,3) NOT NULL,            -- 0.000 to 1.000
    disruption_frequency_24m  DECIMAL(5,3),           -- disruptions per month over 24m
    seasonal_risk_score       DECIMAL(4,3),
    rolling_aqi_4w            DECIMAL(6,2),
    recorded_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Workers

### Table: `worker_profiles`

> Delivery and operational profile — separate from auth (worker_users).

```sql
CREATE TABLE worker_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL UNIQUE REFERENCES worker_users(id),
    home_zone_id    VARCHAR(50) NOT NULL REFERENCES zones(id),
    current_zone_id VARCHAR(50) REFERENCES zones(id),  -- updated from GPS activity
    vehicle_type    VARCHAR(20) NOT NULL,
                    -- two_wheeler | three_wheeler | four_wheeler | bicycle
    preferred_start TIME,
    preferred_end   TIME,
    active_status   VARCHAR(20) NOT NULL DEFAULT 'offline',
                    -- online | offline | on_delivery
    last_seen_at    TIMESTAMP,
    last_gps_lat    DECIMAL(9,6),
    last_gps_lng    DECIMAL(9,6),
    zone_changed_at TIMESTAMP,                         -- for 7-day zone lock enforcement
    total_deliveries INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `worker_zone_history`

> Tracks zone changes for mobility pattern scoring and zone hopping fraud detection.

```sql
CREATE TABLE worker_zone_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    entered_at      TIMESTAMP NOT NULL,
    exited_at       TIMESTAMP,
    entry_reason    VARCHAR(30),                       -- onboarding | gps_drift | manual_update
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Policies & Premiums

### Table: `policies`

> One active policy per worker at any time.

```sql
CREATE TABLE policies (
    id              VARCHAR(20) PRIMARY KEY,          -- pol_<uuid>
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
                    -- active | paused | suspended | cancelled
    coverage_ratio  DECIMAL(4,3) NOT NULL DEFAULT 0.850,
    max_weekly_payout DECIMAL(8,2) NOT NULL,
    enrolled_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    coverage_starts_at TIMESTAMP NOT NULL,            -- enrolled_at + waiting period
    paused_at       TIMESTAMP,
    suspended_at    TIMESTAMP,
    cancelled_at    TIMESTAMP,
    continuity_streak_weeks INTEGER NOT NULL DEFAULT 0,
    missed_weeks_count      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP                                -- soft delete
);
```

---

### Table: `weekly_policy_cycles`

> One record per worker per week. Tracks premium status for that cycle.

```sql
CREATE TABLE weekly_policy_cycles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       VARCHAR(20) NOT NULL REFERENCES policies(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    week_start      DATE NOT NULL,
    week_end        DATE NOT NULL,
    premium_amount  DECIMAL(8,2) NOT NULL,            -- ML-calculated for this week
    premium_paid    BOOLEAN NOT NULL DEFAULT FALSE,
    premium_paid_at TIMESTAMP,
    premium_source  VARCHAR(20),                      -- auto_deduction | manual | advance
    max_payout      DECIMAL(8,2) NOT NULL,
    payout_used     DECIMAL(8,2) NOT NULL DEFAULT 0,
    payout_remaining DECIMAL(8,2),                   -- computed: max_payout - payout_used
    continuity_reward_applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(policy_id, week_start)
);
```

---

### Table: `premium_payments`

> Individual payment transactions for premium collection.

```sql
CREATE TABLE premium_payments (
    id              VARCHAR(20) PRIMARY KEY,          -- pay_<uuid>
    policy_id       VARCHAR(20) NOT NULL REFERENCES policies(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    cycle_id        UUID NOT NULL REFERENCES weekly_policy_cycles(id),
    amount          DECIMAL(8,2) NOT NULL,
    payment_method  VARCHAR(20) NOT NULL,             -- auto_deduction | manual | advance
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                    -- pending | completed | failed
    razorpay_ref    VARCHAR(100),
    paid_at         TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Orders & Earnings

### Table: `orders`

> One record per delivery order assigned to an InDel-enrolled worker.
> Source: Swiggy/Zomato webhook → InDel assigns and tracks.

```sql
CREATE TABLE orders (
    id              VARCHAR(30) PRIMARY KEY,          -- ord_<platform>_<uuid>
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    platform        VARCHAR(20) NOT NULL,             -- swiggy | zomato
    platform_order_id VARCHAR(50),                   -- original platform order ID
    status          VARCHAR(20) NOT NULL DEFAULT 'assigned',
                    -- assigned | picked_up | completed | cancelled
    assigned_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    picked_up_at    TIMESTAMP,
    completed_at    TIMESTAMP,
    cancelled_at    TIMESTAMP,
    cancellation_reason VARCHAR(50),
    earnings        DECIMAL(8,2),                    -- NULL until completed
    distance_km     DECIMAL(6,2),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `earnings_records`

> One record per completed order. Source of truth for income loss calculation.

```sql
CREATE TABLE earnings_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    order_id        VARCHAR(30) NOT NULL REFERENCES orders(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    amount          DECIMAL(8,2) NOT NULL,
    earned_at       TIMESTAMP NOT NULL,              -- order completed_at
    week_start      DATE NOT NULL,                   -- for weekly aggregation
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `weekly_earnings_summaries`

> Aggregated weekly earnings per worker. Generated every Sunday night.
> Used for baseline calculation and premium deduction prompt.

```sql
CREATE TABLE weekly_earnings_summaries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    week_start      DATE NOT NULL,
    week_end        DATE NOT NULL,
    total_earnings  DECIMAL(8,2) NOT NULL DEFAULT 0,
    total_orders    INTEGER NOT NULL DEFAULT 0,
    active_hours    DECIMAL(6,2),
    hourly_rate     DECIMAL(8,2),                    -- total_earnings / active_hours
    platform        VARCHAR(20),
    settled         BOOLEAN NOT NULL DEFAULT FALSE,
    settled_at      TIMESTAMP,                       -- when platform webhook fires
    premium_prompt_sent BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(worker_id, week_start)
);
```

---

### Table: `earnings_baselines`

> 4-week rolling baseline per worker. Recalculated after each weekly summary.
> Used directly by income loss computation.

```sql
CREATE TABLE earnings_baselines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    baseline_hourly_rate    DECIMAL(8,2) NOT NULL,
    baseline_weekly_rate    DECIMAL(8,2) NOT NULL,
    weeks_of_data           INTEGER NOT NULL,        -- 1-4, for cold start awareness
    baseline_source         VARCHAR(30) NOT NULL,
                            -- worker_history | zone_average | peer_group_average
    calculated_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_until     TIMESTAMP NOT NULL,              -- recalculated weekly
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Disruptions

### Table: `disruptions`

> One record per confirmed disruption event per zone.

```sql
CREATE TABLE disruptions (
    id              VARCHAR(20) PRIMARY KEY,          -- dis_<uuid>
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    type            VARCHAR(20) NOT NULL,
                    -- heavy_rain | extreme_heat | severe_aqi | curfew |
                    -- bandh | zone_closure | order_drop | flash_flood
    severity        VARCHAR(10) NOT NULL,             -- low | medium | high | critical
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
                    -- evaluating | confirmed | active | closed | rejected
    confidence_score DECIMAL(4,3) NOT NULL,
    trigger_signals  JSONB NOT NULL,                 -- ["weather_alert", "order_drop_detected"]
    window_start    TIMESTAMP,                       -- set on confirmation
    window_end      TIMESTAMP,                       -- set on close
    duration_hours  DECIMAL(5,2),                    -- computed on close
    eligible_workers_count INTEGER,
    total_claims_generated INTEGER NOT NULL DEFAULT 0,
    total_payout_amount    DECIMAL(10,2) NOT NULL DEFAULT 0,
    external_data   JSONB,                           -- raw weather/AQI API response snapshot
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `disruption_signals`

> Individual signal events that contributed to a disruption confirmation.

```sql
CREATE TABLE disruption_signals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    disruption_id   VARCHAR(20) NOT NULL REFERENCES disruptions(id),
    signal_type     VARCHAR(30) NOT NULL,
                    -- weather_alert | aqi_alert | order_drop | worker_activity_drop
                    -- zone_closure_alert
    source          VARCHAR(50) NOT NULL,            -- openweathermap | openaq | indel_internal
    raw_value       DECIMAL(10,3),                   -- rainfall mm, AQI value, order drop %
    threshold_value DECIMAL(10,3),                   -- what threshold was crossed
    received_at     TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `disruption_eligibility`

> One record per worker per disruption — tracks eligibility evaluation result.

```sql
CREATE TABLE disruption_eligibility (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    disruption_id   VARCHAR(20) NOT NULL REFERENCES disruptions(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    eligible        BOOLEAN NOT NULL,
    ineligibility_reason VARCHAR(50),
                    -- not_active | not_in_zone | below_acceptance_rate |
                    -- allocation_bias | new_enrollment_hold | policy_inactive
    active_before_disruption BOOLEAN,
    logged_in_during_window  BOOLEAN,
    acceptance_rate          DECIMAL(4,3),
    allocation_bias_detected BOOLEAN NOT NULL DEFAULT FALSE,
    evaluated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(disruption_id, worker_id)
);
```

---

## Claims & Payouts

### Table: `claims`

> One record per worker per disruption. Auto-generated after eligibility passes.

```sql
CREATE TABLE claims (
    id              VARCHAR(20) PRIMARY KEY,          -- clm_<uuid>
    disruption_id   VARCHAR(20) NOT NULL REFERENCES disruptions(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    policy_id       VARCHAR(20) NOT NULL REFERENCES policies(id),
    cycle_id        UUID NOT NULL REFERENCES weekly_policy_cycles(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                    -- pending | approved | rejected | fraud_flagged
    disruption_window_start TIMESTAMP NOT NULL,
    disruption_window_end   TIMESTAMP NOT NULL,
    disruption_hours        DECIMAL(5,2) NOT NULL,
    baseline_hourly_rate    DECIMAL(8,2) NOT NULL,
    expected_earnings       DECIMAL(8,2) NOT NULL,
    actual_earnings         DECIMAL(8,2) NOT NULL,
    income_loss             DECIMAL(8,2) NOT NULL,
    coverage_ratio          DECIMAL(4,3) NOT NULL,
    payout_amount           DECIMAL(8,2) NOT NULL,
    cold_start              BOOLEAN NOT NULL DEFAULT FALSE,
    baseline_source         VARCHAR(30) NOT NULL,
    rejection_reason        VARCHAR(100),
    status_reason           VARCHAR(100),             -- explains any status: approved | rejected | fraud_flagged
    auto_generated          BOOLEAN NOT NULL DEFAULT TRUE,
    reviewed_by             VARCHAR(20),              -- insurer_admin_id if manually reviewed
    reviewed_at             TIMESTAMP,
    created_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMP,               -- soft delete
    UNIQUE(disruption_id, worker_id)               -- one claim per worker per disruption
);
```

---

### Table: `claim_fraud_scores`

> ML fraud scoring output stored for every claim — full audit trail.

```sql
CREATE TABLE claim_fraud_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id        VARCHAR(20) NOT NULL REFERENCES claims(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),

    -- Layer 1: Isolation Forest
    isolation_forest_score  DECIMAL(4,3),
    isolation_forest_verdict VARCHAR(20),            -- clean | suspicious | fraud

    -- Layer 2: DBSCAN
    dbscan_cluster_id       INTEGER,                 -- -1 = noise point (flagged)
    dbscan_verdict          VARCHAR(20),             -- in_cluster | noise_point | insufficient_data

    -- Layer 3: Rule Overlay
    rule_violations         JSONB,                   -- ["gps_zone_mismatch", "orders_during_disruption"]
    hard_reject             BOOLEAN NOT NULL DEFAULT FALSE,

    -- Combined
    overall_fraud_score     DECIMAL(4,3) NOT NULL,
    verdict                 VARCHAR(10) NOT NULL,    -- clean | review | fraud
    routing                 VARCHAR(20) NOT NULL,    -- auto_approve | delayed | manual_review
    explanation             TEXT,

    -- Input features snapshot for audit
    input_features          JSONB NOT NULL,

    scored_at               TIMESTAMP NOT NULL DEFAULT NOW(),
    model_version           VARCHAR(20),
    created_at              TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `payouts`

> One record per approved payout attempt. Retry creates new record.

```sql
CREATE TABLE payouts (
    id              VARCHAR(20) PRIMARY KEY,          -- pay_<uuid>
    claim_id        VARCHAR(20) NOT NULL REFERENCES claims(id),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    amount          DECIMAL(8,2) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'queued',
                    -- queued | processing | credited | failed | retrying
    payment_method  VARCHAR(20) NOT NULL DEFAULT 'upi',
                    -- upi | wallet | bank_transfer
    upi_id          VARCHAR(100),
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    razorpay_ref    VARCHAR(100),
    razorpay_status VARCHAR(50),
    attempt_count   INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    retry_schedule  JSONB,                           -- {"next_retry": "...", "backoff_seconds": 60}
    credited_at     TIMESTAMP,
    failure_reason  VARCHAR(200),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `maintenance_checks`

> Self-service claim audit triggered by worker from dashboard.
> Max 3 per worker per day enforced at application layer.

```sql
CREATE TABLE maintenance_checks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    claim_id        VARCHAR(20) REFERENCES claims(id),
    disruption_id   VARCHAR(20) REFERENCES disruptions(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                    -- pending | ai_responded | reviewer_responded | closed
    worker_query    TEXT,
    ai_response     TEXT,
    ai_response_lang VARCHAR(10),
    ai_shap_explanation JSONB,                       -- raw SHAP breakdown before templating
    ai_plain_language   TEXT,                        -- templated plain language version
    reviewer_id     VARCHAR(20) REFERENCES insurer_admin_users(id),
    reviewer_response TEXT,
    reviewer_responded_at TIMESTAMP,
    triggered_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    ai_responded_at TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## ML Model Outputs

### Table: `premium_model_outputs`

> Full XGBoost + SHAP output stored per worker per weekly cycle.

```sql
CREATE TABLE premium_model_outputs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    cycle_week_start DATE NOT NULL,
    risk_score      DECIMAL(4,3) NOT NULL,
    weekly_premium  DECIMAL(8,2) NOT NULL,
    shap_breakdown  JSONB NOT NULL,
    -- {
    --   "flood_risk_zone": 6.00,
    --   "rolling_aqi_pattern": 3.00,
    --   "income_instability": 2.00,
    --   "base_rate": 7.00,
    --   "seasonal_adjustment": 4.00
    -- }
    plain_language  TEXT NOT NULL,
    input_features  JSONB NOT NULL,                  -- full feature vector for audit
    model_version   VARCHAR(20),
    calculated_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(worker_id, cycle_week_start)
);
```

---

### Table: `forecast_model_outputs`

> Prophet forecast output stored per zone per weekly run.

```sql
CREATE TABLE forecast_model_outputs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id         VARCHAR(50) NOT NULL REFERENCES zones(id),
    forecast_run_at TIMESTAMP NOT NULL,
    forecast_week_start DATE NOT NULL,
    forecast_week_end   DATE NOT NULL,
    daily_forecasts JSONB NOT NULL,
    -- [
    --   {"date": "2026-03-22", "claim_probability": 0.31,
    --    "expected_claims": 4, "expected_payout": 2200.00},
    --   ...
    -- ]
    week_aggregate  JSONB NOT NULL,
    -- {"expected_claims": 56, "expected_payout": 30800.00,
    --  "recommended_reserve": 46200.00}
    model_version   VARCHAR(20),
    limitation_note TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(zone_id, forecast_week_start)
);
```

---

## Notifications

### Table: `notifications`

> All worker-facing notifications. Delivered via FCM.

```sql
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id       VARCHAR(20) NOT NULL REFERENCES worker_users(id),
    type            VARCHAR(40) NOT NULL,
                    -- disruption_alert | claim_generated | payout_credited |
                    -- premium_due | premium_prompt | policy_paused |
                    -- continuity_reward | maintenance_check_response
    title           VARCHAR(100) NOT NULL,
    body            TEXT NOT NULL,
    language        VARCHAR(10) NOT NULL DEFAULT 'en',
    metadata        JSONB,                           -- disruption_id, claim_id, payout_id etc
    read            BOOLEAN NOT NULL DEFAULT FALSE,
    sent_via_fcm    BOOLEAN NOT NULL DEFAULT FALSE,
    fcm_message_id  VARCHAR(200),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    read_at         TIMESTAMP
);
```

---

## Audit & System

### Table: `idempotency_keys`

> Stores used idempotency keys for payment deduplication.

```sql
CREATE TABLE idempotency_keys (
    key             VARCHAR(100) PRIMARY KEY,
    response_body   JSONB NOT NULL,
    status_code     INTEGER NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `kafka_event_log`

> Audit log of all Kafka events published by the system.

```sql
CREATE TABLE kafka_event_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic           VARCHAR(100) NOT NULL,
    event_type      VARCHAR(50) NOT NULL,
    payload         JSONB NOT NULL,
    partition       INTEGER,
    offset_value    BIGINT,
    published_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### Table: `api_request_log`

> Request log for debugging and rate limit enforcement.

```sql
CREATE TABLE api_request_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      VARCHAR(50) NOT NULL,
    gateway         VARCHAR(20) NOT NULL,            -- worker | insurer | platform
    method          VARCHAR(10) NOT NULL,
    path            VARCHAR(200) NOT NULL,
    user_id         VARCHAR(20),
    status_code     INTEGER,
    duration_ms     INTEGER,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Indexes

```sql
-- Worker lookups
CREATE INDEX idx_worker_users_phone ON worker_users(phone);
CREATE INDEX idx_worker_profiles_home_zone ON worker_profiles(home_zone_id);
CREATE INDEX idx_worker_profiles_current_zone ON worker_profiles(current_zone_id);
CREATE INDEX idx_worker_profiles_active_status ON worker_profiles(active_status);

-- Policy lookups
CREATE INDEX idx_policies_worker ON policies(worker_id);
CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_weekly_cycles_worker_week ON weekly_policy_cycles(worker_id, week_start);

-- Order & earnings lookups
CREATE INDEX idx_orders_worker ON orders(worker_id);
CREATE INDEX idx_orders_zone ON orders(zone_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_completed_at ON orders(completed_at);
CREATE INDEX idx_earnings_records_worker_week ON earnings_records(worker_id, week_start);
CREATE INDEX idx_weekly_summaries_worker ON weekly_earnings_summaries(worker_id);
CREATE INDEX idx_earnings_baselines_worker ON earnings_baselines(worker_id);

-- Disruption lookups
CREATE INDEX idx_disruptions_zone ON disruptions(zone_id);
CREATE INDEX idx_disruptions_status ON disruptions(status);
CREATE INDEX idx_disruptions_window ON disruptions(window_start, window_end);
CREATE INDEX idx_disruption_eligibility_disruption ON disruption_eligibility(disruption_id);
CREATE INDEX idx_disruption_eligibility_worker ON disruption_eligibility(worker_id);

-- Claims lookups
CREATE INDEX idx_claims_worker ON claims(worker_id);
CREATE INDEX idx_claims_disruption ON claims(disruption_id);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_claims_policy ON claims(policy_id);
CREATE INDEX idx_fraud_scores_claim ON claim_fraud_scores(claim_id);
CREATE INDEX idx_fraud_scores_verdict ON claim_fraud_scores(verdict);

-- Payout lookups
CREATE INDEX idx_payouts_claim ON payouts(claim_id);
CREATE INDEX idx_payouts_worker ON payouts(worker_id);
CREATE INDEX idx_payouts_status ON payouts(status);
CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);

-- ML output lookups
CREATE INDEX idx_premium_outputs_worker_week ON premium_model_outputs(worker_id, cycle_week_start);
CREATE INDEX idx_forecast_outputs_zone_week ON forecast_model_outputs(zone_id, forecast_week_start);

-- Notification lookups
CREATE INDEX idx_notifications_worker ON notifications(worker_id);
CREATE INDEX idx_notifications_unread ON notifications(worker_id, read) WHERE read = FALSE;

-- Zone history
CREATE INDEX idx_zone_risk_history_zone ON zone_risk_history(zone_id, recorded_at DESC);
CREATE INDEX idx_worker_zone_history_worker ON worker_zone_history(worker_id, entered_at DESC);

-- Kafka log
CREATE INDEX idx_kafka_log_topic ON kafka_event_log(topic, published_at DESC);

-- Soft delete (partial indexes — only index non-deleted rows)
CREATE INDEX idx_worker_users_active ON worker_users(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_policies_active ON policies(worker_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_claims_active ON claims(worker_id) WHERE deleted_at IS NULL;

-- Auth
CREATE INDEX idx_auth_tokens_user ON auth_tokens(user_id, user_type);
CREATE INDEX idx_auth_tokens_expires ON auth_tokens(expires_at) WHERE revoked = FALSE;
```

---

## Entity Relationship Summary

| Table | Relates To | Relationship |
|---|---|---|
| `worker_users` | `worker_profiles` | 1:1 |
| `worker_users` | `policies` | 1:many (one active at a time) |
| `worker_users` | `orders` | 1:many |
| `worker_users` | `earnings_records` | 1:many |
| `worker_users` | `claims` | 1:many |
| `worker_users` | `notifications` | 1:many |
| `zones` | `worker_profiles` | 1:many (home zone) |
| `zones` | `disruptions` | 1:many |
| `zones` | `zone_risk_history` | 1:many |
| `policies` | `weekly_policy_cycles` | 1:many |
| `weekly_policy_cycles` | `premium_payments` | 1:many |
| `orders` | `earnings_records` | 1:1 |
| `disruptions` | `disruption_signals` | 1:many |
| `disruptions` | `disruption_eligibility` | 1:many |
| `disruptions` | `claims` | 1:many (one per eligible worker) |
| `claims` | `claim_fraud_scores` | 1:1 |
| `claims` | `payouts` | 1:many (retries) |
| `claims` | `maintenance_checks` | 1:many |
| `worker_users` | `premium_model_outputs` | 1:many (one per week) |
| `zones` | `forecast_model_outputs` | 1:many (one per week) |

---

## Seed Data Required

On first deploy, seed the following before any worker can onboard:

```sql
-- Zones (sample — expand to full list)
INSERT INTO zones (id, display_name, city, state, lat_center, lng_center,
                   risk_profile, base_premium, max_payout) VALUES
('tambaram_chennai',      'Tambaram, Chennai',      'Chennai',   'Tamil Nadu',  12.9249, 80.1000, 'high',   22.00, 800.00),
('koramangala_bengaluru', 'Koramangala, Bengaluru',  'Bengaluru', 'Karnataka',   12.9352, 77.6245, 'medium', 16.00, 700.00),
('rohini_delhi',          'Rohini, Delhi',           'Delhi',     'Delhi',       28.7041, 77.1025, 'medium', 17.00, 700.00),
('kothrud_pune',          'Kothrud, Pune',           'Pune',      'Maharashtra', 18.5074, 73.8077, 'low',    11.00, 600.00);
```

---

*Database Design v1 — Team ImaginAI — Guidewire DEVTrails 2026*
*Note: Schema will evolve during implementation. Run migrations via golang-migrate — never modify tables manually in any environment.*