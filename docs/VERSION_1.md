# InDel — Phase 2 Build Plan
## Version 1 — Automation & Protection

> Theme: "Protect Your Worker"
> Timeline: March 21 – April 4 (2 weeks)
> Hackathon Requirement: Registration, Policy Management, Dynamic Premium, Claims Management

---

## What This Phase Delivers

A working end-to-end income protection system demonstrable in 2 minutes.

A judge watching the demo will see:

1. A delivery worker registers on InDel and enrolls in income protection
2. A flood event is detected in their zone — automatically
3. Their earnings drop is calculated against their baseline
4. A claim is generated without the worker doing anything
5. A payout lands in their UPI wallet
6. The insurer dashboard updates in real time — loss ratio, fraud queue, pool health

No manual claims. No paperwork. No phone calls. Zero touch from worker to payout.

That is what Version 1 delivers.

---

## What We Are NOT Building in This Phase

This is as important as what we are building.

| Out of Scope | Reason |
|---|---|
| DBSCAN fraud layer | Isolation Forest is sufficient for V1 demo |
| Maintenance Check AI feature | Time permitting — if Week 2 finishes early |
| Transit disruption edge cases | Time permitting — if Week 2 finishes early |
| Interstate travel rules | Time permitting — if Week 2 finishes early |
| Full SHAP plain language in worker app | Backend stores it, frontend shows simplified version |
| iOS app | Android only for V1 |
| Real Aadhaar/PAN KYC | Mocked — button exists, always returns verified |
| gRPC between Go and Python | REST is sufficient at demo scale |
| DeepAR forecasting | Prophet is the V1 model |
| AWS deployment | Render free tier for V1 |
| Real Swiggy/Zomato API | Webhook simulation via demo endpoints |

If it is not in the build list below, it does not get built in Phase 2.

---

## Build List — Week 1 (March 21–28)

### Priority 1 — Foundation (Days 1–2)

**Goal: One command brings the entire stack up.**

- [ ] Docker Compose file — Go backend, Python ML services, PostgreSQL, Kafka, all in one
- [ ] golang-migrate setup — migration files for all tables in correct dependency order
- [ ] Seed script — 4 zones (Tambaram Chennai, Koramangala Bengaluru, Rohini Delhi, Kothrud Pune)
- [ ] Environment config — `.env.example` with all required keys documented
- [ ] GitHub Actions CI — runs on every push, builds all Docker images, fails fast

**Done when:** `docker-compose up` brings everything up cleanly on any team member's machine.

---

### Priority 2 — Auth & Worker Registration (Days 2–3)

**Hackathon requirement: Registration Process**

- [ ] `POST /api/v1/auth/otp/send` — Firebase Phone OTP
- [ ] `POST /api/v1/auth/otp/verify` — verify OTP, return JWT
- [ ] `POST /api/v1/worker/onboard` — create `worker_users` + `worker_profiles` record
- [ ] `GET /api/v1/worker/profile` — fetch profile
- [ ] JWT middleware on all authenticated routes
- [ ] Kotlin Android: OTP screen → profile setup screen → home screen

**Done when:** A worker can register with a phone number, verify OTP, complete their profile, and land on a home screen showing their zone and coverage status.

---

### Priority 3 — Policy & Premium (Days 3–5)

**Hackathon requirement: Insurance Policy Management + Dynamic Premium Calculation**

- [ ] XGBoost model trained on synthetic data — IMD historical weather + simulated zone disruption logs
- [ ] SHAP breakdown computed per worker per zone
- [ ] `POST /ml/v1/premium/calculate` — risk score → weekly premium
- [ ] `premium_model_outputs` stored for every calculation
- [ ] `POST /api/v1/worker/policy/enroll` — creates `policies` + `weekly_policy_cycles` record
- [ ] `GET /api/v1/worker/policy` — returns active policy with current week premium
- [ ] `POST /api/v1/worker/policy/premium/pay` — manual payment via Razorpay sandbox
- [ ] Insurer dashboard: premium rates by zone visible
- [ ] Kotlin Android: coverage enrollment screen, premium display, pay button

**Done when:** A worker in Tambaram Chennai sees a ₹22 premium and a worker in Kothrud Pune sees ₹11 — different because the ML model produced different risk scores for their zones. Worker can enroll and pay.

---

### Priority 4 — Orders & Earnings (Days 5–7)

- [ ] `POST /api/v1/platform/webhooks/order/assigned` — creates `orders` record
- [ ] `POST /api/v1/platform/webhooks/order/completed` — updates order, creates `earnings_records`
- [ ] Weekly earnings summary generation — aggregates `earnings_records` into `weekly_earnings_summaries`
- [ ] Baseline calculation — 4-week rolling hourly rate into `earnings_baselines`
- [ ] Cold start handling — zone average substitution for workers with < 20 deliveries
- [ ] `GET /api/v1/worker/earnings` — returns this week vs baseline
- [ ] Kotlin Android: earnings screen showing actual vs protected income

**Done when:** Simulated orders flow through webhooks, earnings accumulate, and a worker's hourly baseline is calculable.

---

## Build List — Week 2 (March 28 – April 4)

### Priority 5 — Disruption Detection (Days 8–9)

**Hackathon requirement: 3–5 automated triggers**

Triggers to implement (all 5 for full marks):

| Trigger | API | Threshold |
|---|---|---|
| Heavy rain | OpenWeatherMap | Rainfall > 50mm in 2 hours |
| Extreme heat | OpenWeatherMap | Temperature > 42°C during active hours |
| Severe AQI | OpenAQ | AQI > 300 (hazardous) |
| Curfew / zone closure | Mock API (simulated) | Zone closure alert received |
| Platform order drop | InDel internal | Order volume drops > 40% vs rolling average |

- [ ] OpenWeatherMap poller — polls every 10 minutes per active zone, publishes to `indel.weather.alerts` Kafka topic
- [ ] OpenAQ poller — polls every 30 minutes per zone, publishes to `indel.aqi.alerts`
- [ ] Order drop detector — sliding window on `orders` table, publishes to `indel.zone.order-drop`
- [ ] Zone closure simulator — mock endpoint that fires `ZONE_CLOSURE_ALERT` for demo
- [ ] Disruption evaluator — consumes all four topics, computes confidence score
- [ ] Multi-signal validation — both external signal AND internal order drop required
- [ ] `disruptions` record created on confirmation
- [ ] `disruption_signals` records created per contributing signal
- [ ] `POST /api/v1/demo/trigger-disruption` — demo endpoint bypasses API wait
- [ ] Kotlin Android: disruption alert notification via FCM

**Done when:** A simulated flood event in Tambaram Chennai creates a confirmed disruption record with confidence score > threshold, and the worker's phone receives a push notification.

---

### Priority 6 — Claims Pipeline (Days 9–11)

**Hackathon requirement: Claims Management + Zero-touch claim process**

- [ ] Worker eligibility evaluation — was active before disruption, logged in during window, acceptance rate check
- [ ] Income loss computation — baseline hourly rate × disruption hours − actual earnings
- [ ] Payout calculation — income loss × coverage ratio, capped at weekly maximum
- [ ] Isolation Forest fraud scoring — trained on synthetic normal claim patterns
- [ ] Rule overlay — GPS zone check, completed orders during disruption check
- [ ] Confidence-based routing — auto approve / delayed / manual review
- [ ] `claims` record auto-generated — worker does nothing
- [ ] `claim_fraud_scores` stored — full audit trail
- [ ] `POST /internal/v1/claims/:claim_id/payout` — queues to `indel.payouts.queued` Kafka topic
- [ ] Payout processor — consumes queue, calls Razorpay sandbox, updates `payouts` record
- [ ] Worker notification on payout credited — FCM push
- [ ] `GET /api/v1/worker/claims` — claim history in worker app
- [ ] Kotlin Android: claims screen, payout status, amount breakdown

**Done when:** From disruption confirmed to payout credited happens automatically with zero worker action. Worker sees a notification and opens the app to find money already credited.

---

### Priority 7 — Insurer Dashboard (Days 11–13)

**This is InDel's differentiator. Build it properly.**

- [ ] `GET /api/v1/insurer/overview` — KPI cards: active workers, loss ratio, pending claims, reserve
- [ ] `GET /api/v1/insurer/loss-ratio` — by city and zone
- [ ] `GET /api/v1/insurer/claims/fraud-queue` — flagged claims with fraud signals visible
- [ ] `GET /api/v1/insurer/forecast` — Prophet 7-day claim probability per zone
- [ ] `GET /api/v1/insurer/pool/health` — premiums collected vs paid out this week
- [ ] Prophet model trained on synthetic historical disruption data per zone
- [ ] `forecast_model_outputs` stored weekly
- [ ] Vite + React + Tremor: KPI cards, loss ratio chart by zone, fraud queue table, 7-day forecast chart, reserve recommendation panel

**Done when:** An insurer logging into the dashboard sees live loss ratios by zone, a fraud queue with claim details, and a 7-day forecast telling them how much reserve to hold.

---

### Priority 8 — Demo Preparation (Days 13–14)

- [ ] Seed database with realistic synthetic data — 3 workers, 2 zones, 4 weeks of order history
- [ ] End-to-end demo scenario rehearsed — flood trigger → eligibility → claim → payout → dashboard update
- [ ] `POST /api/v1/demo/trigger-disruption` tested and reliable
- [ ] `POST /api/v1/demo/settle-earnings` tested — triggers premium prompt in worker app
- [ ] `POST /api/v1/demo/reset-zone` tested — resets for repeat demo runs
- [ ] Backup demo video recorded — in case live demo has connectivity issues
- [ ] Render deployment stable — all services running, health endpoints green
- [ ] 2-minute demo video recorded and uploaded to public link

**Done when:** The demo runs end-to-end in under 2 minutes without any manual intervention beyond hitting the demo trigger endpoint.

---

## The 2-Minute Demo Script

This is the exact flow to demonstrate. Every second is allocated.

```
0:00 – 0:20  Problem statement
             "Delivery workers lose 20-30% income during disruptions.
              No protection exists. We built InDel."

0:20 – 0:45  Worker registration + policy enrollment
             Show onboarding screen, OTP verification,
             zone assignment (Tambaram Chennai), premium shown as ₹22
             "The ML model priced this zone higher because of monsoon history."

0:45 – 1:05  Trigger the disruption
             Hit demo endpoint — flood confirmed in Tambaram Chennai
             Worker phone receives FCM notification: "Disruption detected in your zone"

1:05 – 1:25  Zero-touch claim
             Show claims screen — claim auto-generated, fraud check passed,
             payout ₹527 credited to UPI
             Worker did nothing. No form. No call. No waiting.

1:25 – 1:50  Insurer dashboard
             Switch to dashboard — loss ratio updated, claim visible in pipeline,
             fraud queue shows 0 flags, 7-day forecast shows next disruption probability
             "This is what the insurer sees. Real economics. Real data."

1:50 – 2:00  Close
             "Others verify presence. We verify income.
              Data → Trust → Usage."
```

---

## What Version 1 Delivers — Summary

| Deliverable | Status |
|---|---|
| Worker registration with OTP | Built |
| Dynamic premium via XGBoost + SHAP | Built |
| 5 automated disruption triggers | Built |
| Zero-touch claim generation | Built |
| Isolation Forest fraud detection | Built |
| Automated UPI payout via Razorpay sandbox | Built |
| Insurer dashboard — loss ratio, fraud queue, forecast | Built |
| Worker Android app — registration, policy, claims, earnings | Built |
| 2-minute demo video | Recorded |
| Demo endpoints for reliable live demo | Built |

---

## What Version 1 Proves to Judges

**To a Guidewire engineer:**
The architecture is production-grade. Kafka for async payouts, first-party data ownership, ML audit trail, idempotent payments. This isn't a hackathon toy.

**To an insurance professional:**
The financial model is correct. Loss ratios are real, reserve logic exists, reinsurance layer is planned. The insurer dashboard speaks their language.

**To anyone:**
A delivery worker in Chennai during a flood gets paid automatically before they even know a claim was filed. That's the product.

---

*Phase 2 Build Plan — Team ImaginAI — Guidewire DEVTrails 2026*
*Freeze this scope. Build exactly this. Nothing more.*