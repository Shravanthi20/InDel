# InDel — Team Assignments & GitHub Rules
## Phase 2 Version 1 — March 21 to April 4

---

## Branch Strategy — Trunk Based

```
main (protected — always deployable)
  └── feat/<scope>/<short-description>
```

No develop branch. No release branches. Feature branches are short-lived — opened, reviewed, merged, deleted. Main deploys on every merge.

---

## GitHub Rules

### Branch naming
```
feat/auth/otp-firebase
feat/ml/xgboost-premium
feat/disruption/weather-poller
feat/android/onboarding-screen
feat/dashboard/loss-ratio-chart
fix/auth/jwt-expiry-bug
chore/infra/docker-compose-setup
```

Format: `type/scope/what`
Types: `feat` `fix` `chore` `refactor` `test` `docs`

### Commit messages — Conventional Commits
```
feat(auth): add Firebase OTP send endpoint
feat(ml): train XGBoost premium model with synthetic data
fix(kafka): resolve consumer group rebalance on restart
chore(docker): add healthcheck to postgres service
refactor(claims): extract income loss calculation to service layer
test(fraud): add isolation forest smoke tests
docs(api): update worker onboard request schema
```

Format: `type(scope): short description in sentence case`
- Present tense — "add" not "added"
- No capital first letter
- No full stop at end
- Under 72 characters

### PR rules
- Every change goes through a PR — no direct pushes to main, ever
- PR title follows conventional commit format
- PR description must include: what changed, why, how to test it
- Minimum 1 approval before merge
- All CI checks must pass before merge
- Delete branch after merge — always
- Draft PRs allowed for early feedback — prefix title with `[WIP]`
- Keep PRs small — one feature or one fix per PR
- Review others' PRs the same day they are opened — 2-week sprint, no delays

### What never goes in a PR
- `.env` files with real secrets
- `google-services.json`
- `firebase-credentials.json`
- Any API keys, tokens, or passwords
- Generated files (`*.pkl`, large CSVs, `node_modules`, `.gradle`)

### Protected branch rules (set in GitHub repo settings)
- `main` requires PR before merging
- `main` requires at least 1 approving review
- `main` requires status checks to pass (CI workflow)
- `main` does not allow force pushes
- `main` does not allow deletion

### Always do this before starting a new branch
```bash
git checkout main
git pull
git checkout -b feat/your-scope/your-description
```

---

## Effort Weights

Before assigning, every task was weighted honestly:

| Task | Effort |
|---|---|
| Layer 0 — Foundation, Docker, migrations | Medium |
| Layer 1 — Auth, JWT, Firebase OTP, worker models | Medium-Heavy |
| Track A — Policy logic + XGBoost + SHAP ML service | Heavy |
| Track B — Orders, earnings, baseline, cold start | Medium-Heavy |
| Track C — 5 disruption triggers, Kafka, pollers, consumers | Heavy |
| Track D — Android app, 6 screens, Compose, FCM, Hilt | Heavy |
| Layer 3 — Claims pipeline, eligibility, fraud ML, payout | Very Heavy |
| Layer 4 — Insurer dashboard, Prophet ML, Go backend, React | Heavy |
| Layer 4 — Platform dashboard | Light |
| Layer 5 — Seed data, E2E testing, Render deployment | Medium |

Total effort split across 5 members — each carries roughly 2 heavy items across the 2 weeks.

---

## Assignments

---

### Member 1 — Shravanthi Satyanarayanan
**Tracks: Layer 1 (Auth) + Track A (Policy + Premium ML)**
**Effort: Medium-Heavy + Heavy = Heavy**

Layer 1 starts Day 1. Track A starts Day 3 immediately after auth merges.

```
Layer 1 — Auth & Worker Registration (Days 1–3)
Branch: feat/auth/worker-registration

Backend (Go):
- backend/pkg/jwt/jwt.go
- backend/pkg/firebase/otp.go
- backend/pkg/response/response.go
- backend/internal/config/config.go
- backend/internal/database/postgres.go
- backend/internal/database/migrate.go
- backend/internal/middleware/auth.go
- backend/internal/middleware/rbac.go
- backend/internal/models/worker_user.go
- backend/internal/models/worker_profile.go
- backend/internal/models/auth_token.go
- backend/internal/models/zone.go
- backend/internal/handlers/worker/auth.go
- backend/internal/handlers/worker/onboarding.go
- backend/internal/services/auth_service.go
- backend/internal/services/worker_service.go
- backend/internal/router/worker_router.go (auth + onboarding routes only)
- backend/cmd/worker-gateway/main.go
- GET /api/v1/health
- GET /api/v1/status

Endpoints:
POST /api/v1/auth/otp/send
POST /api/v1/auth/otp/verify
POST /api/v1/worker/onboard
GET  /api/v1/worker/profile
```

```
Track A — Policy + Premium ML (Days 3–7)
Branch: feat/policy/enrollment-and-premium

ML service (Python):
- ml/premium/main.py
- ml/premium/features.py
- ml/premium/model.py
- ml/premium/shap_explainer.py
- ml/premium/train.py
- ml/premium/data/synthetic_training_data.csv
- ml/premium/Dockerfile

Backend (Go):
- backend/internal/models/policy.go
- backend/internal/models/weekly_policy_cycle.go
- backend/internal/models/premium_payment.go
- backend/internal/models/premium_model_output.go
- backend/internal/handlers/worker/policy.go
- backend/internal/handlers/worker/premium.go
- backend/internal/services/policy_service.go
- backend/internal/services/premium_service.go
- backend/pkg/idempotency/idempotency.go
- backend/pkg/razorpay/razorpay.go

Endpoints:
POST /api/v1/worker/policy/enroll
GET  /api/v1/worker/policy
PUT  /api/v1/worker/policy/pause
PUT  /api/v1/worker/policy/cancel
GET  /api/v1/worker/policy/premium
POST /api/v1/worker/policy/premium/pay
POST /ml/v1/premium/calculate (internal)
```

**Done when:** Worker in Tambaram Chennai sees ₹22, worker in Kothrud Pune sees ₹11. Different because ML produced different risk scores. Worker can enroll and pay.

---

### Member 2 — Saravana Priyaa C R
**Tracks: Layer 0 (Foundation) + Track C (Disruption + Kafka) + Demo Prep**
**Effort: Medium + Heavy + Medium = Heavy**

Layer 0 is solo and must finish Day 1. Track C starts Day 2 in parallel with others.

```
Layer 0 — Foundation (Day 1)
Branch: chore/infra/foundation-setup

- docker-compose.yml
    postgres, kafka, zookeeper, all Go services, all ML services
- docker-compose.demo.yml
    same with INDEL_ENV=demo and pre-seeded data
- .env.example — every variable documented
- .gitignore
- migrations/ — all 9 migration files in correct dependency order
- scripts/seed.sql — 4 zones seeded
- scripts/reset-demo.sh — drops and re-seeds demo data
- .github/workflows/ci.yml — builds all Docker images on push, fails fast
- .github/workflows/deploy.yml — deploys to Render on merge to main
```

```
Track C — Disruption Triggers + Kafka Pipeline (Days 2–9)
Branch: feat/disruption/triggers-and-kafka

Backend (Go):
- backend/internal/kafka/producer.go (Sarama)
- backend/internal/kafka/consumer.go
- backend/internal/kafka/topics.go
- backend/internal/models/disruption.go
- backend/internal/models/disruption_signal.go
- backend/internal/models/disruption_eligibility.go
- backend/internal/pollers/weather_poller.go (OpenWeatherMap, every 10 min)
- backend/internal/pollers/aqi_poller.go (OpenAQ, every 30 min)
- backend/internal/workers/weather_consumer.go
- backend/internal/workers/aqi_consumer.go
- backend/internal/workers/order_drop_consumer.go
- backend/internal/services/disruption_service.go
    confidence score, multi-signal validation, window open/close
- backend/internal/handlers/demo/demo.go
- backend/internal/router/demo_router.go

5 triggers to implement:
1. heavy_rain — OpenWeatherMap rainfall > 50mm / 2 hours
2. extreme_heat — temperature > 42C during active hours
3. severe_aqi — OpenAQ AQI > 300
4. curfew — mock zone closure via demo endpoint
5. order_drop — internal: volume drops > 40% vs rolling average

Kafka topics to set up:
- indel.weather.alerts
- indel.aqi.alerts
- indel.zone.order-drop
- indel.disruption.confirmed

Demo endpoints:
POST /api/v1/demo/trigger-disruption
POST /api/v1/demo/settle-earnings
POST /api/v1/demo/reset-zone
```

**Done when:** Simulated flood in Tambaram Chennai creates a confirmed disruption record. Demo endpoint fires the full internal pipeline reliably. All 5 trigger types produce confirmed disruption records.

```
Demo Prep (Days 12–14)
Branch: chore/demo/seed-and-e2e

- scripts/generate-synthetic-data.py
    3 workers (Tambaram Chennai, Koramangala Bengaluru, Rohini Delhi)
    4 weeks order history per worker, realistic earnings variance
- Seed database with synthetic data
- Render deployment — all services running, all health endpoints green
- docker-compose.demo.yml tested end to end
- Run end-to-end demo scenario 5 times
    flood trigger → eligibility → claim → payout → dashboard update
- Verify demo endpoints reliable across all 5 runs
- Verify reset-demo.sh resets cleanly between runs
```

---

### Member 3 — Subikha MV
**Tracks: Track B (Earnings Engine) + Layer 3 (Claims Pipeline)**
**Effort: Medium-Heavy + Very Heavy = Very Heavy**

Track B starts Day 3. Layer 3 starts Day 9 when Tracks A, B, C are all merged.
Layer 3 is the heaviest single task in the entire build — offset by Track B being lighter.

```
Track B — Orders, Earnings & Baseline Engine (Days 3–8)
Branch: feat/earnings/orders-and-baseline

Backend (Go):
- backend/internal/models/order.go
- backend/internal/models/earnings_record.go
- backend/internal/models/weekly_earnings_summary.go
- backend/internal/models/earnings_baseline.go
- backend/internal/handlers/worker/earnings.go
- backend/internal/handlers/platform/webhooks.go
- backend/internal/services/earnings_service.go
    per-order recording, weekly aggregation,
    4-week baseline calculation, cold start handling
- backend/internal/router/platform_router.go

Endpoints:
POST /api/v1/platform/webhooks/order/assigned
POST /api/v1/platform/webhooks/order/completed
POST /api/v1/platform/webhooks/order/cancelled
POST /api/v1/platform/webhooks/earnings/settled
GET  /api/v1/worker/earnings
GET  /api/v1/worker/earnings/history
GET  /api/v1/worker/earnings/baseline
```

```
Layer 3 — Claims Pipeline (Days 9–12)
Branch: feat/claims/pipeline

ML service (Python):
- ml/fraud/main.py
- ml/fraud/isolation_forest.py (Layer 1)
- ml/fraud/rules.py (Layer 3 hard disqualifiers)
- ml/fraud/scorer.py (combines layers → verdict + routing)
- ml/fraud/train.py
- ml/fraud/Dockerfile

Backend (Go):
- backend/internal/models/claim.go
- backend/internal/models/claim_fraud_score.go
- backend/internal/models/payout.go
- backend/internal/models/notification.go
- backend/internal/services/eligibility_service.go
    was worker active before disruption, logged in during window,
    acceptance rate, allocation bias check
- backend/internal/services/income_loss_service.go
    baseline × hours − actual, coverage ratio, weekly cap, cold start
- backend/internal/services/claim_service.go
    auto-generate for eligible workers, call fraud service, route
- backend/internal/services/fraud_service.go
- backend/internal/services/payout_service.go
    Razorpay sandbox, Kafka queue, retry on failure
- backend/internal/services/notification_service.go (FCM payout_credited)
- backend/internal/workers/disruption_consumer.go
- backend/internal/workers/claim_consumer.go
- backend/internal/workers/fraud_consumer.go
- backend/internal/workers/payout_consumer.go
- backend/internal/handlers/worker/claims.go
- backend/internal/handlers/worker/payouts.go

Kafka topics to consume:
- indel.disruption.confirmed → eligibility evaluation
- indel.claims.generated → fraud scoring
- indel.claims.scored → routing
- indel.payouts.queued → Razorpay payout
- indel.payouts.failed → retry with exponential backoff

Endpoints:
GET  /api/v1/worker/claims
GET  /api/v1/worker/claims/:claim_id
GET  /api/v1/worker/wallet
GET  /api/v1/worker/payouts
POST /ml/v1/fraud/score (internal)
```

**Done when:** Disruption confirmed → eligibility evaluated → income loss calculated → fraud check passed → claim auto-generated → payout credited to Razorpay sandbox → worker receives FCM push. Zero worker action.

---

### Member 4 — Gayathri U
**Tracks: Track D (Android Worker App) + Platform Dashboard**
**Effort: Heavy + Light = Heavy**

Android starts Day 2 (Android Studio setup). Platform dashboard fills remaining time in Week 2.

```
Track D — Android Worker App (Days 2–11)
Branch: feat/android/worker-app

Setup (Day 2):
Create Android project in Android Studio
  Name: InDel
  Package: com.imaginai.indel
  Language: Kotlin
  Min SDK: API 26
  Template: Empty Activity
  Build: Kotlin DSL

Add to app/build.gradle.kts:
  Retrofit2 + OkHttp (API calls)
  Hilt (dependency injection)
  Jetpack Compose (UI)
  Navigation Compose
  ViewModel + LiveData
  DataStore Preferences
  Firebase BOM (Auth + Messaging)
  Kotlin Coroutines

Foundation (Days 2–3):
- InDelApplication.kt — Hilt application class
- di/AppModule.kt — dependency injection
- ui/theme/Theme.kt, Color.kt, Type.kt
- ui/navigation/NavGraph.kt
- data/api/ApiClient.kt — Retrofit + JWT interceptor
- data/api/AuthApiService.kt
- data/api/WorkerApiService.kt
- data/local/PreferencesDataStore.kt
- service/InDelFirebaseMessagingService.kt

Screens (Days 3–10):
- ui/auth/OtpScreen.kt + OtpViewModel.kt
    phone number entry, OTP entry, JWT stored in DataStore
- ui/auth/OnboardingScreen.kt + OnboardingViewModel.kt
    name, zone picker, vehicle type, UPI ID
- ui/home/HomeScreen.kt + HomeViewModel.kt
    coverage status badge, zone name, earnings strip,
    disruption alert card when active disruption in zone
- ui/policy/PolicyScreen.kt + PolicyViewModel.kt
    active policy details, premium with SHAP breakdown,
    enrollment CTA for unenrolled workers
- ui/policy/PremiumPayScreen.kt
- ui/earnings/EarningsScreen.kt + EarningsViewModel.kt
    this week actual vs baseline, protected income, weekly history
- ui/claims/ClaimsScreen.kt + ClaimsViewModel.kt
- ui/claims/ClaimDetailScreen.kt
    payout breakdown, disruption window, income loss calc, fraud verdict

Data layer:
- data/model/ — data classes matching all API response schemas
- data/repository/ — Auth, Worker, Policy, Earnings, Claims repositories
```

```
Platform Dashboard (Days 11–13)
Branch: feat/dashboard/platform

Backend (Go):
- backend/internal/handlers/platform/workers.go
- backend/internal/handlers/platform/zones.go
- backend/internal/handlers/platform/analytics.go
- backend/cmd/platform-gateway/main.go

Frontend (Vite + React):
- Full platform-dashboard/src/ structure
- Pages: Login, Overview, Workers, Zones, Analytics
- Components: Sidebar, TopBar, ZoneCard
```

**Done when:** Full Android flow works on device — OTP login, onboarding, home with coverage status, policy with premium, earnings, claims. Platform dashboard shows active workers and zone disruptions.

---

### Member 5 — Rithanya K A
**Tracks: Insurer Dashboard (Go backend + Prophet ML + React frontend)**
**Effort: Heavy**

Insurer dashboard starts Day 9 when claims data exists.

```
Insurer Dashboard (Days 9–13)
Branch: feat/dashboard/insurer

ML service (Python):
- ml/forecast/main.py
- ml/forecast/prophet_model.py (per-zone train + predict)
- ml/forecast/features.py
- ml/forecast/train.py
- ml/forecast/Dockerfile

Backend (Go):
- backend/internal/models/forecast_model_output.go
- backend/internal/services/forecast_service.go
- backend/internal/handlers/insurer/overview.go
- backend/internal/handlers/insurer/loss_ratio.go
- backend/internal/handlers/insurer/claims.go (pipeline + fraud queue)
- backend/internal/handlers/insurer/forecast.go
- backend/internal/handlers/insurer/premiums.go
- backend/internal/handlers/insurer/workers.go
- backend/internal/handlers/insurer/maintenance_checks.go
- backend/internal/router/insurer_router.go
- backend/cmd/insurer-gateway/main.go

Endpoints:
GET /api/v1/insurer/overview
GET /api/v1/insurer/pool/health
GET /api/v1/insurer/loss-ratio
GET /api/v1/insurer/loss-ratio/history
GET /api/v1/insurer/claims
GET /api/v1/insurer/claims/fraud-queue
PUT /api/v1/insurer/claims/:claim_id/approve
PUT /api/v1/insurer/claims/:claim_id/reject
GET /api/v1/insurer/forecast
GET /api/v1/insurer/reserve
POST /ml/v1/forecast/zone (internal)

Frontend (Vite + React + Tremor):
- insurer-dashboard/src/api/client.ts
- insurer-dashboard/src/api/insurer.ts
- insurer-dashboard/src/types/index.ts
- insurer-dashboard/src/pages/Login.tsx
- insurer-dashboard/src/pages/Overview.tsx
    KPI cards: active workers, loss ratio, pending claims, reserve
    Pool health bar: premiums collected vs paid out
- insurer-dashboard/src/pages/LossRatio.tsx
    Recharts bar by zone and city
    Benchmark line at 65%
- insurer-dashboard/src/pages/FraudQueue.tsx
    Claims table with fraud signals, approve/reject buttons
- insurer-dashboard/src/pages/Forecast.tsx
    7-day Recharts line chart + reserve recommendation
- insurer-dashboard/src/components/layout/Sidebar.tsx + TopBar.tsx
- insurer-dashboard/src/components/cards/KpiCard.tsx
- insurer-dashboard/src/components/cards/ReserveCard.tsx
- insurer-dashboard/src/components/charts/LossRatioChart.tsx
- insurer-dashboard/src/components/charts/ForecastChart.tsx
- insurer-dashboard/src/components/tables/FraudQueueTable.tsx
```

**Done when:** Insurer dashboard shows live loss ratios by zone, fraud queue with signals, 7-day forecast with reserve recommendation.

---

## Effort Balance Check

| Member | Tasks | Estimated Effort |
|---|---|---|
| Shravanthi | Auth + Policy + Premium ML | Heavy |
| Saravana Priyaa | Foundation + Disruption + Kafka | Heavy |
| Subikha | Earnings Engine + Claims Pipeline | Heavy |
| Gayathri | Android App + Platform Dashboard | Heavy |
| Rithanya | Insurer Dashboard (full stack) | Heavy |

No one person is carrying more than one very heavy track. Claims pipeline (Subikha) is the heaviest single item — offset by Track B being the lightest backend track. Insurer dashboard (Rithanya) covers three layers (ML + Go + React) but each layer is contained.

---

## Dependency — What Blocks What

```
Layer 0 (Saravana Priyaa Day 1)
    ↓ unblocks everyone
Layer 1 Auth (Shravanthi Days 1–3)
    ↓ unblocks all four parallel tracks
    ┌───────────────┬───────────────┬───────────────┐
Track A            Track B         Track C         Track D
Policy + ML        Earnings        Disruption      Android
(Shravanthi)       (Subikha)       (Saravana)      (Gayathri)
Days 3–7           Days 3–8        Days 2–9        Days 2–11
    └───────────────┴───────────────┘
              ↓ all three merged
         Layer 3 Claims Pipeline
              (Subikha Days 9–12)
                    ↓
    ┌───────────────────────────────┐
Insurer Dashboard              Platform Dashboard
(Rithanya Days 9–13)           (Gayathri Days 11–13)
    └───────────────────────────────┘
                    ↓
              Demo Prep
          (all members, Days 12–14)
```

---

## Week-by-Week View

### Week 1 — March 21–28

| Day | Saravana Priyaa | Shravanthi | Gayathri | Subikha | Rithanya |
|---|---|---|---|---|---|
| 1 | Layer 0 foundation | Waiting — reviewing docs | Android Studio setup | Android Studio setup | Dashboard research |
| 2 | Start Kafka setup | Auth — JWT + OTP | Android foundation + Hilt | Waiting for auth | Insurer dashboard scaffolding |
| 3 | Kafka producers | Auth complete → policy enroll | Auth + onboarding screens | Start orders + webhooks | Wait for claims data |
| 4 | Weather + AQI pollers | XGBoost model training | Home + policy screens | Order completion + earnings | Insurer API client setup |
| 5 | Disruption service | SHAP explainer | Premium pay screen | Weekly summary aggregation | Forecast ML research |
| 6 | Order drop detector | Policy handlers + routes | Earnings screen | Baseline calculation | Prophet model setup |
| 7 | Disruption confidence scoring | Premium service complete → PR | Claims + detail screens | Cold start handling | Forecast service scaffolding |

### Week 2 — March 28 – April 4

| Day | Saravana Priyaa | Shravanthi | Gayathri | Subikha | Rithanya |
|---|---|---|---|---|---|
| 8 | Disruption consumers + demo endpoints | Assist claims if needed | FCM push handler | Earnings baseline complete → PR | Prophet train + predict |
| 9 | Disruption complete → PR | Review PRs | Platform dashboard start | Start claims — eligibility service | Start insurer backend |
| 10 | Render deployment setup | Synthetic data generation | Platform workers + zones pages | Income loss + claim auto-generation | Loss ratio endpoints + chart |
| 11 | CI/CD pipeline stable | ML model final training | Android final polish + PR | Fraud ML service | Fraud queue endpoint + table |
| 12 | E2E testing support | E2E verification | Platform dashboard PR | Payout + Kafka consumers → PR | Forecast endpoint + chart |
| 13 | Seed data + demo reset script | Review all open PRs | Review open PRs | E2E demo scenario testing | Insurer dashboard complete → PR |
| 14 | Demo rehearsal + E2E | Demo rehearsal + ML verify | Demo rehearsal + Android | Demo rehearsal + E2E | Demo rehearsal + dashboard |

---

## Demo Prep — All Members (Days 12–14)

Branch: `chore/demo/seed-and-e2e`

```
Saravana Priyaa:
- scripts/generate-synthetic-data.py
    3 workers (Tambaram Chennai, Koramangala Bengaluru, Rohini Delhi)
    4 weeks order history per worker, realistic earnings variance
- Seed database before demo run
- Render deployment stable, all health endpoints green
- docker-compose.demo.yml tested end to end

Shravanthi:
- Train all three ML models on synthetic data
- Verify SHAP breakdown outputs correct plain language
- Verify Isolation Forest scores synthetic clean claims correctly

Subikha:
- Run end-to-end demo scenario 5 times
    flood trigger → eligibility → claim → payout → dashboard update
- Verify all demo endpoints reliable
- Verify FCM push delivers on physical device

Gayathri:
- Android app tested on physical device
- All screens polished and demo flow walkthrough rehearsed

Rithanya:
- Verify loss ratio numbers match README unit economics
- Verify SHAP breakdowns display correctly in insurer dashboard
- Insurer dashboard demo walkthrough rehearsed
```

**Done when:** End-to-end demo runs cleanly 5 times without failure. Every member has rehearsed their part.

---

## Key Rules to Share With Team

```
1.  Never push directly to main. Ever.
2.  Always pull latest main before starting a new branch.
    git checkout main && git pull && git checkout -b feat/your-scope/description
3.  Keep branches short-lived. Open a PR within 2 days of starting work.
4.  PR titles follow conventional commits format.
5.  All CI checks must pass before requesting review.
6.  Review others' PRs the same day they are opened — 2-week sprint.
7.  If you are blocked, say so immediately — do not sit on it for a day.
8.  If a task is taking longer than expected, cut scope — do not miss the merge window.
9.  Merge and delete the branch. Do not let stale branches accumulate.
10. Demo endpoints only exist when INDEL_ENV=demo.
    Never put demo shortcuts inside production logic.
```

---

*Team Assignments v1 — ImaginAI — Guidewire DEVTrails 2026*
