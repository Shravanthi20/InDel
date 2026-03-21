# InDel — Tech Stack

> Every technology choice below was made for a specific reason. This document explains what we chose, why we chose it, and what we deliberately decided not to use.

---

## Guiding Principle

> Choose reliable and familiar technology over impressive and complex technology. Ship a working core loop in 4 weeks. Every unfamiliar tool is a risk.

---

## Stack Overview

| Layer | Technology |
|---|---|
| Worker frontend | Kotlin (Android) |
| Insurer dashboard | Vite + React + Tremor |
| Backend | Go |
| ML microservices | Python + FastAPI |
| Event pipeline | Apache Kafka (disruption flow only) |
| Backend ↔ ML comms | REST (gRPC planned) |
| Database | PostgreSQL |
| ORM + Migrations | GORM + golang-migrate |
| Auth | JWT + Firebase Phone OTP |
| Weather trigger | OpenWeatherMap API |
| AQI trigger | OpenAQ API |
| Payment simulation | Razorpay Sandbox |
| Containerisation | Docker + Docker Compose |
| CI/CD | GitHub Actions + Render |

---

## Layer-by-Layer Breakdown

---

### Worker Frontend — Kotlin (Android)


**What:** Native Android application for delivery worker onboarding, coverage status, disruption alerts, claim history, and payout notifications.

**Why Kotlin:**
- Workers primarily use Android devices in the Indian market — iOS is a secondary priority addressed in a later phase
- Native Android gives direct access to GPS, background location services, and push notifications without a bridge layer
- Kotlin is concise, null-safe, and the current standard for Android development
- Native performance matters for workers on budget devices (₹8k–₹15k range) where React Native or Flutter can introduce lag

**Why Android only first:**
- Delivery workers in India are overwhelmingly on Android
- A single native codebase ships faster and more reliably than a cross-platform one under a 4-week constraint
- iOS support is planned for a subsequent phase once the Android experience is stable

**What we're not building yet:**
- iOS version
- Kotlin Multiplatform
- Web worker view

---

### Insurer Dashboard — Vite + React + Tremor

**What:** Internal analytics dashboard for insurance providers — live loss ratio by zone, premium pool health, fraud queue, 7-day disruption forecast, reserve recommendation.

**Why Vite + React:**
- The team has prior React experience — familiarity beats optimisation under a tight timeline
- Vite's build speed is significantly faster than Create React App or Next.js for a pure SPA use case
- The insurer dashboard has no SEO requirements, no public-facing pages, and low concurrent users — Next.js server-side rendering is unnecessary overhead for this profile
- Pure SPA is the right architecture for a real-time data dashboard

**Why Tremor:**
- Component library built specifically for analytics dashboards and internal tools
- Ships KPI cards, loss ratio charts, fraud queue tables, and status badges out of the box
- Looks professional without custom CSS work
- Lets the frontend focus on data wiring rather than component design

**Why not Next.js:**
- Next.js is optimised for public-facing applications with SEO, dynamic routing, and server-side rendering — none of which the insurer dashboard needs
- Adds build complexity without adding value for an internal tool with a handful of users
- Vite + React achieves the same result faster with less configuration

**Why not Retool:**
- Retool would have been the fastest path to a working dashboard
- Rejected because it limits customisation of the loss ratio visualisations and fraud queue logic that are central to InDel's insurer-facing differentiator

---

### Backend — Go

**What:** Core API server handling business logic — worker management, policy lifecycle, disruption event processing, claim orchestration, payout coordination.

**Why Go:**
- High concurrency with goroutines — critical for handling simultaneous disruption events across multiple zones without blocking
- Strongly typed, fast compilation, low memory footprint
- Well-suited for event-driven architectures and Kafka consumer/producer patterns
- Prior production Go experience on the team (federated social networking application with AI moderators) — no learning curve risk
- Standard library is comprehensive enough that minimal third-party dependencies are needed, reducing integration risk during a sprint

**Kafka client:** Sarama (pure Go, no CGO dependency — avoids Docker build issues in CI/CD pipeline)

**ORM:** GORM for database interactions, golang-migrate for schema versioning

**Why not Node.js / FastAPI for the backend:**
- Go's concurrency model is a better fit for the event-driven disruption pipeline than Node's event loop
- Python is reserved for ML microservices where the scientific ecosystem (scikit-learn, XGBoost, Prophet) is unmatched — mixing Python into the core backend would create unnecessary service boundary confusion

---

### ML Microservices — Python + FastAPI

**What:** Three independent ML services — premium calculation, fraud detection, disruption forecasting — each exposed as a REST endpoint consumed by the Go backend.

**Why Python:**
- scikit-learn, XGBoost, SHAP, Prophet, and DBSCAN are all Python-native
- No equivalent exists in Go or Kotlin for this ML stack
- Python is the correct language for this layer — not a compromise

**Why FastAPI:**
- Async, high performance, auto-generates OpenAPI docs
- Pydantic validation ensures clean data contracts between Go backend and ML services
- Lightweight enough to run three separate microservice instances without significant resource overhead

**The three ML services:**

| Service | Model | Purpose | Build priority |
|---|---|---|---|
| Premium calculator | XGBoost + SHAP | Risk score → weekly premium | Week 1 |
| Fraud detector | Isolation Forest + DBSCAN + Rules | Claim legitimacy scoring | Week 1 (Layer 1 only) |
| Disruption forecaster | Prophet | 7-day zone claim probability for insurer dashboard | Week 2 |

**Why not a single ML monolith:**
- Each model has a different retraining cadence (monthly, weekly, weekly)
- Independent services can be updated, retrained, and redeployed without touching the others
- Isolation Forest alone is sufficient for the prototype fraud layer — DBSCAN is added Week 3 if time allows

---

### Event Pipeline — Apache Kafka

**What:** Async event streaming for the disruption detection pipeline specifically.

**Why Kafka:**
- Log-based architecture provides durable event replay — critical for reprocessing payouts after a payment gateway failure during a mass disruption event
- Persistent offset model gives a complete audit trail of every payout attempt — a regulatory expectation for insurance products
- Horizontal scalability without blocking the main claim pipeline during simultaneous mass disruption events

**Why not RabbitMQ:**
- RabbitMQ's queue-deletion model makes audit replay harder to guarantee
- For an insurance product where every payout must be traceable, Kafka's immutable log is not optional

**Scope — Kafka is used only for:**
- WEATHER_ALERT events → zone disruption detection pipeline
- ORDER_DROP_DETECTED events → zone aggregation
- Async payout queue → Razorpay sandbox disbursement

**REST is used for everything else:**
- Go backend → Python ML microservices (premium calculation, fraud scoring)
- Frontend → Go backend
- Point-to-point queries that are request-response by nature

**gRPC (planned):**
- If REST latency between Go and Python ML services becomes a bottleneck under load, gRPC will replace REST for those specific calls in a later phase
- Not in scope for the 4-week prototype

---

### Database — PostgreSQL

**What:** Primary relational database for workers, policies, claims, earnings records, zone data, and event logs.

**Why PostgreSQL:**
- ACID-compliant — financial ledger data requires transaction integrity
- PostGIS extension available for zone-level geographic queries if needed
- Mature, well-documented, reliable under the team's expected load
- Direct integration with GORM on the Go side and SQLAlchemy on the Python side

**Schema versioning:** golang-migrate keeps the schema versioned and team-synced from Day 1, preventing environment drift during a multi-developer sprint.

---

### Auth & Identity — JWT + Firebase Phone OTP

**What:** Role-based access control for three user types — worker, insurer admin, platform admin.

**Why JWT:**
- Stateless, works cleanly with Go backend
- Role claims embedded in token — no database round trip for permission checks
- Standard, well-understood, no third-party dependency

**Why Firebase Phone OTP:**
- Worker onboarding requires phone number verification
- Firebase handles Indian phone numbers reliably on the free tier
- No custom OTP infrastructure to build or maintain

**KYC (Aadhaar / PAN):**
- Mocked for the prototype
- Real KYC integration requires insurer partnership and IRDAI compliance infrastructure
- Explicitly out of scope for the hackathon build

---

### External APIs

| API | Purpose | Tier |
|---|---|---|
| OpenWeatherMap | Rainfall, temperature, flood threshold triggers | Free tier — sufficient for prototype |
| OpenAQ | AQI disruption triggers by zone | Free tier |
| Razorpay Sandbox | UPI payout simulation | Test mode — no real money |
| IMD API | Secondary weather oracle, government alerts | Add Week 2 if OpenWeatherMap needs backup |
| Traffic / Zone Closure API | Curfew, strike detection | Mocked for demo — real integration is post-hackathon |

---

### Infrastructure & DevOps

**Docker + Docker Compose**

Every service — Go backend, Python ML microservices, PostgreSQL, Kafka, React dashboard — runs in a container from Day 1.

Why this matters: if containers are not set up from the start, Week 4 becomes a debugging marathon when one team member's environment differs from another's. One command (`docker-compose up`) brings the entire stack up identically on every machine.

---

**GitHub Actions — CI/CD**

Pipeline runs on every push to any branch:

```yaml
on: [push]

jobs:
  test-backend:
    - Go unit tests
    - go vet + golint

  test-ml:
    - pytest for ML microservices
    - Model smoke tests (does Isolation Forest score a claim?)

  build:
    - Docker build for all services
    - Fails fast if any service doesn't containerise cleanly

  deploy:
    - On merge to main only
    - Auto-deploy to Render
```

---

**Render — Hosting**

- Free tier handles Python ML microservices, Go backend, and Vite dashboard without AWS complexity
- One-click deploy from GitHub
- Move to AWS EC2 / RDS in Week 4 only if Render free tier limits become a problem under demo load

---

## What We Deliberately Decided Not to Build

| Technology | Reason skipped |
|---|---|
| React Native / Flutter | Worker app is Android-native Kotlin — no cross-platform overhead |
| Next.js | SSR is unnecessary for an internal insurer dashboard with low concurrent users |
| Retool | Limits customisation of the loss ratio and fraud queue views that are core differentiators |
| DeepAR / TFT | Needs 6+ months of real zone data to outperform Prophet — post-hackathon upgrade |
| DBSCAN (Week 1) | Isolation Forest covers Layer 1 fraud detection — DBSCAN added Week 3 if time allows |
| gRPC (Week 1) | REST is sufficient at hackathon scale — gRPC added if latency becomes a bottleneck |
| Aadhaar / PAN KYC | Requires insurer partnership and IRDAI compliance infrastructure — mocked for prototype |
| Traffic / Zone Closure API | Mocked for demo — real integration is post-hackathon scope |
| AWS (Week 1) | Render free tier is sufficient — AWS added only if needed in Week 4 |
| iOS (Kotlin Multiplatform) | Android-first, iOS in subsequent phase |

---

## Build Priority by Week

| Week | Focus |
|---|---|
| Week 1 | Docker Compose setup, Go backend skeleton, PostgreSQL schema, Kafka pipeline, Isolation Forest fraud service, OpenWeatherMap trigger, Razorpay sandbox payout, basic worker onboarding |
| Week 2 | XGBoost premium calculator, Prophet forecasting, income loss calculation engine, zone aggregation, insurer dashboard data wiring |
| Week 3 | Insurer dashboard polish (loss ratio, fraud queue, reserve recommendation), DBSCAN fraud layer, end-to-end demo scenario seeded with synthetic data |
| Week 4 | Demo rehearsal, backup video recording, synthetic data refinement, load testing, Render deployment stability |

---

*Tech stack maintained by Team ImaginAI — Guidewire DEVTrails 2026*
*Note: This is the planned stack. Technologies may change during development as implementation constraints become clear.*