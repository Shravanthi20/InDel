# InDel — Insure, Deliver

> An AI-powered parametric income insurance platform for gig delivery workers — combining delivery management and automated income protection into a single integrated system.

**Team:** ImaginAI
**Hackathon:** Guidewire DEVTrails 2026
**Persona:** E-commerce
**Current Phase:** Phase 1 (Ideation)

> **Note:** Specific values such as premium amounts, payout figures, coverage ratios, and trigger thresholds are illustrative estimates for design and modelling purposes only. Final figures will be refined during development in collaboration with relevant stakeholders. API integrations and third-party service references are subject to change.

---

## Table of Contents

1. [The Problem](#the-problem)
2. [What We Plan to Build](#what-we-plan-to-build)
3. [Stakeholders](#stakeholders)
4. [System Architecture](#system-architecture)
5. [End-to-End Pipeline](#end-to-end-pipeline)
6. [How the System Works — Step by Step](#how-the-system-works--step-by-step)
7. [AI and ML Models](#ai-and-ml-models)
8. [Weekly Premium Model](#weekly-premium-model)
9. [Fraud Prevention — Economic Activity Consistency Defense](#fraud-prevention--economic-activity-consistency-defense)
10. [Risk Controls and Edge Cases](#risk-controls-and-edge-cases)
11. [Illustrative Unit Economics](#illustrative-unit-economics)
12. [Scenario Walkthroughs](#scenario-walkthroughs)
13. [Dashboards](#dashboards)
14. [Compliance and Regulatory Considerations](#compliance-and-regulatory-considerations)
15. [Tech Stack](#tech-stack)
16. [Team ImaginAI](#team-imaginai)

---

## The Problem

India's gig delivery workers earn based on completed orders. When external disruptions occur — heavy rain, extreme heat, severe pollution, curfews, strikes, or sudden order drops — deliveries stop and income falls sharply. Workers can lose 20–30% of monthly earnings during these events with no financial protection in place.

**Valid claim triggers under the problem scope:**

| Category | Disruption Types |
|---|---|
| Environmental | Extreme heat, heavy rain, floods, flash floods, severe pollution |
| Social | Curfews, strikes, zone closures |
| Platform | Significant order volume drop in zone |

Traditional insurance covers accidents, vehicles, and health. Income lost because conditions made it impossible to work falls entirely outside the scope of any product available to these workers.

Existing parametric insurance attempts face a structural problem: they depend on third-party delivery platforms to share worker activity data. Those platforms have no incentive to share this data, making verification unreliable and fraud detection weak.

---

## What We Plan to Build

InDel is a **B2B platform built for insurance providers** that combines a delivery management system and an automated parametric income protection engine into a single integrated product.

The insurance provider is the primary customer. InDel gives them a ready-to-deploy infrastructure that handles delivery worker management, real-time disruption monitoring, income loss calculation, claim verification, and payout processing — all in one system, without depending on third-party delivery app data.

InDel owns the delivery operations side too: workers receive assignments through the platform, complete deliveries, and earn — just like any delivery management platform. The insurance layer runs in the background using the same first-party activity data the delivery system already collects.

```
Traditional Approach:
Insurance Provider → needs data from Swiggy/Zomato → API access unlikely → incomplete data → weak fraud detection

InDel Approach:
Insurance Provider deploys InDel → delivery system + insurance engine share one data layer → accurate verification → reliable payouts
```

---

## Stakeholders

**Insurance Provider (Primary B2B Customer)**
Purchases and deploys InDel. Gets access to a previously uninsured worker segment with an integrated data pipeline, automated claim processing, and risk analytics — without negotiating data agreements with existing delivery platforms.

**Delivery Platform Partner (e.g. Swiggy, Zomato)**
InDel is not replacing consumer delivery apps. InDel handles the worker management and insurance layer; the platform partner benefits from having their delivery workers covered and financially protected.

**Delivery Worker (End Beneficiary)**
Uses InDel for delivery assignments. Can opt into income protection at onboarding or any point thereafter. Coverage runs in the background based on actual activity — no active management required.

---

## System Architecture

```
+----------------------------------------------------------+
|                     InDel Platform                       |
|                                                          |
|  +------------------+     +-------------------------+   |
|  |  Delivery Engine |     |    Insurance Engine     |   |
|  |                  |     |                         |   |
|  | Order Allocation |<--->| Policy Management       |   |
|  | Worker Tracking  |<--->| Premium Calculation     |   |
|  | GPS Activity     |<--->| Disruption Detection    |   |
|  | Earnings Records |<--->| Claim Processing        |   |
|  +------------------+     +-------------------------+   |
|           |                          |                   |
|           +----------+  +-----------+                   |
|                      |  |                               |
|              +--------+--+--------+                     |
|              |    AI / ML Engine  |                     |
|              |                    |                     |
|              | Risk Scoring       |                     |
|              | Fraud Detection    |                     |
|              | Disruption Forecast|                     |
|              +--------------------+                     |
|                                                         |
|  +----------------------------------------------------+ |
|  |              External Data Integrations            | |
|  | OpenWeatherMap | OpenAQ | Traffic API | UPI/Payment| |
|  +----------------------------------------------------+ |
+----------------------------------------------------------+
```

---

## End-to-End Pipeline

InDel operates as an **event-driven**, zone-based parametric insurance platform. Rather than continuously polling data sources, the system reacts to structured events generated by external services and internal platform activity. This makes it more efficient, scalable, and resilient to API delays.

```
Onboarding & Policy Initialization
        ↓
Event-Driven Data Ingestion
(WEATHER_ALERT / AQI_ALERT / ORDER_DROP_DETECTED / WORKER_ACTIVITY_UPDATE)
        ↓
Zone Aggregation & Monitoring
(sliding time windows — rolling order volume, active worker counts)
        ↓
Disruption Detection + Confidence Scoring
(multi-signal: environmental + order drop + worker activity)
        ↓
Multi-Signal Validation
(external signal + internal order drop both required)
        ↓
Disruption Window Creation
(start = first valid trigger, end = signals return to normal)
        ↓
Worker Eligibility Evaluation
(active before + during disruption, acceptance rate threshold)
        ↓
Fairness Validation
(platform allocation bias check)
        ↓
Income Loss Computation
(baseline vs actual earnings during disruption window)
        ↓
Behavioral Validation
        ↓
Fraud Detection + Risk-Based Routing
(Isolation Forest + DBSCAN + Identity checks)
        ↓
Automated Claim Generation + Payout Guardrails
        ↓
Hybrid Approval + Worker Notification
        ↓
Asynchronous Payout Processing
(queue-based: Kafka / RabbitMQ — UPI / Wallet / Bank)
        ↓
Zone Risk Update + AI/ML Feedback Loop
```

---

## How the System Works — Step by Step

### Step 1 — Worker Onboarding and Enrollment

Workers register on InDel as delivery partners. Onboarding collects:

- Name, location, home zone
- Preferred working hours
- Bank account / UPI ID for payouts
- Delivery vehicle type
- Device ID (for identity linking and fraud prevention)

Income protection enrollment is **optional** and presented as a separate choice. Workers who decline can still use the platform and enroll later from their dashboard. Coverage starts from the following weekly cycle on enrollment.

**Edge cases covered:**
- Multiple accounts detected via device + bank + phone identity linking
- Fake onboarding mitigated through KYC validation

---

### Step 2 — Event-Driven Data Ingestion

Instead of continuously polling data sources, the system uses an **event-driven architecture**. Multiple data sources generate structured real-time events that feed into the platform:

| Event Type | Source | Trigger |
|---|---|---|
| `WEATHER_ALERT` | OpenWeatherMap | Rainfall / temperature / flood threshold crossed |
| `AQI_ALERT` | OpenAQ / WAQI | Pollution level exceeds safe limits |
| `ORDER_DROP_DETECTED` | InDel internal | Zone-level order flow anomaly |
| `WORKER_ACTIVITY_UPDATE` | InDel platform | Login, acceptance, completion events |
| `ZONE_CLOSURE_ALERT` | Traffic API / Govt alerts | Curfew, strike, or zone restriction detected |

Each event is structured and typed. The system buffers incoming events to handle API delays and falls back to internal platform signals when external APIs are unavailable.

**Edge cases covered:**
- API failure → fallback to InDel internal order and activity signals
- Delayed API responses → buffered time windows with lag tolerance

---

### Step 3 — Zone Aggregation and Monitoring

All incoming events are aggregated at the **zone level** rather than the individual worker level — a deliberate design choice for scalability and noise reduction.

For each zone, the system maintains rolling metrics using **sliding time windows** (e.g., last 30–60 minutes):

- Current order volume vs historical average
- Current active workers vs historical average

This zone-level aggregation allows the system to detect localized disruptions while smoothing out short-term fluctuations. Individual worker data is only evaluated once a zone-level disruption has been confirmed.

---

### Step 4 — Disruption Detection and Confidence Scoring

When a triggering event is received, the system generates a disruption candidate for the corresponding zone. Candidate triggers include:

| Disruption Type | Indicative Trigger | Source |
|---|---|---|
| Heavy Rain | Rainfall above threshold in worker's zone | OpenWeatherMap |
| Extreme Heat | Temperature above threshold during active hours | OpenWeatherMap |
| Severe Pollution | AQI above hazardous level | OpenAQ / WAQI |
| Curfew / Bandh | Verified zone closure | Traffic API / Govt alert |
| Platform Order Drop | Significant order volume drop in zone | InDel internal |
| Flash Flood | Flood alert issued by IMD | Weather alert API |

A **Trigger Confidence Score** is then computed from multiple signals:

- Environmental signals (weather, AQI)
- Order volume drop (InDel internal)
- Worker activity drop

A disruption is considered valid only if the confidence score exceeds a predefined threshold. Below threshold, the event is ignored or placed under continued monitoring. This prevents false positives from noisy or incomplete data.

**Multi-Signal Validation:** A disruption is confirmed only when both an external signal (weather/AQI/curfew) **and** an internal signal (zone-level order drop) are simultaneously present. Environmental anomalies with no observable delivery impact do not trigger payouts. A time-lag window accounts for cases where environmental changes and operational impact are not perfectly synchronized.

**Edge cases covered:**
- False positives from weak or noisy signals suppressed by confidence threshold
- Delayed impact captured via time-lag window
- Multiple triggers in the same zone merged into a single disruption event

---

### Step 5 — Disruption Window Management

Once a disruption is confirmed, the system creates a **disruption window** for the affected zone:

- **Start time:** Defined as the first valid trigger crossing the confidence threshold
- **End time:** Defined as the point at which all signals return to normal levels

This window represents the period during which workers may have experienced income loss. The system supports both short-duration (micro) disruptions and longer disruptions, with duration caps applied to prevent excessive exposure per event.

---

### Step 6 — Worker Eligibility Evaluation and Fairness Validation

For each worker in the affected zone, eligibility requires:

- Worker was active on InDel before the disruption began
- Worker was logged in during the disruption window
- Worker's order acceptance rate is above a defined threshold

**Fairness Validation:** If a worker's orders during the disruption are significantly below the zone average due to apparent allocation bias (platform under-assigning orders to them regardless of disruption), this is treated as a platform issue rather than disruption impact and excluded from claim calculation.

**Edge cases covered:**
- Idle workers excluded from payouts
- Fake participation filtered
- Platform allocation bias prevented from generating false claims

---

### Step 7 — Income Loss Computation

```
Baseline hourly rate   = Average hourly earnings over past 4 weeks (InDel data)
Disruption window      = Time from disruption start to resolution
Expected earnings      = Baseline hourly rate × disruption hours
Actual earnings        = Earnings recorded in InDel during the disruption window
Income loss            = Expected earnings − Actual earnings
Payout amount          = Income loss × coverage ratio (capped at weekly maximum)
```

**Cold Start Handling:** For new workers without a 4-week history, baseline is derived from zone average or peer group average for the same zone and vehicle type.

**Illustrative example:**

A worker earns Rs. 120/hour on average. A flood event is logged from 11:40 AM to 5:30 PM (~5h 50m).

```
Expected earnings:       Rs. 120 × 5.83 hrs = Rs. 700
Actual earnings:         2 partial deliveries = Rs. 80
Estimated income loss:   Rs. 620
Estimated payout:        Rs. 558 (at ~90% coverage ratio)
```

**Edge cases covered:**
- New workers without earnings history (cold start)
- Partial disruptions pay proportional loss
- No income drop → no payout

---

### Step 8 — Behavioral Validation and Fraud Detection

**Behavioral Validation** confirms:

- Work pattern consistency during the disruption window
- Session continuity and active delivery attempts
- Similarity with other workers in the same zone during the same event

**Fraud Detection Stack (three independent layers):**

**Layer 1 — Isolation Forest (Anomaly Detection)**
Trained on expected claim behavior patterns. Flags workers whose claim profile deviates statistically from the zone-wide cluster.

Input features:
- GPS trail consistency during disruption window
- Ratio of claimed loss to historical earnings baseline
- Claim frequency over rolling 8-week window
- Zone-wide claim clustering (are others in the zone also claiming?)
- Mobility pattern score (how stable is the worker's operating zone)

**Layer 2 — DBSCAN Cluster-Based Behavior Analysis**
Workers in the same zone during the same disruption should show similar claim patterns. DBSCAN clusters workers by behavior per disruption event. Workers who fall outside any cluster (noise points) are flagged — this catches fraud that Isolation Forest alone misses: a worker whose individual claim history looks normal but whose behavior diverges from everyone else during a specific event.

**Layer 3 — Rule Overlay (Hard Disqualifiers)**
- Worker GPS not in affected zone at trigger time → auto-reject
- InDel platform shows completed deliveries during claimed disruption window → auto-reject
- Anomaly score above threshold from Layer 1 → route to manual review queue

All three layers are independent. A claim must clear all three to proceed.

**Confidence-Based Routing:**
- Low-risk claims → auto-approved
- Medium-risk claims → delayed for additional validation
- High-risk claims → manual review queue

---

### Step 9 — Automated Claim Generation and Payout Guardrails

A claim is automatically generated when:

- Disruption is confirmed as valid
- Worker is eligible
- Income loss exceeds a minimum threshold
- All fraud checks pass

**Payout Guardrails:**
- Maximum payout per worker per week
- Maximum payout per disruption event
- No duplicate payouts for overlapping disruptions (merged into single window)

Final payout is capped accordingly.

---

### Step 10 — Hybrid Approval and Asynchronous Payout Processing

The system pre-approves the payout. The worker is notified with:

- Disruption event details
- Income loss calculation breakdown
- Final payout amount

Payout is either auto-credited after a short delay or optionally confirmed by the worker.

All payouts are processed **asynchronously** via a queue-based system (e.g., Kafka or RabbitMQ). This enables:

- Batch processing of large volumes during mass disruption events
- Automatic retry on payment failure
- Horizontal scalability without blocking the main claim pipeline

Payment is sent via:

- UPI direct transfer
- InDel in-app wallet (can offset future premiums)
- Bank transfer (next-day settlement for amounts above a defined threshold)

Workers receive transparent communication on the full payout breakdown.

**Edge cases covered:**
- Payment failures handled via retry logic within the async queue
- High-volume simultaneous payout scenarios handled through batching

---

### Step 11 — Zone Risk Update and AI/ML Feedback Loop

After each disruption event:

- Zone risk score is updated
- Disruption frequency for the zone is updated
- Claim density is updated

All three ML models are retrained using new data from the event. A temporal adjustment layer accounts for delay between the disruption trigger and observable order drop impact.

---

### Step 12 — Maintenance Check (Self-Service Claim Audit)

Workers who believe they should have been eligible for a claim but were not notified or had their eligibility check fail can trigger a Maintenance Check from their dashboard.

**Phase 1 — Automated AI Response:**

1. System sends a query to the AI API with the worker's activity data, GPS records, zone disruption signals, and SHAP-based explanation of the eligibility model's assessment.
2. Response is returned in the worker's preferred language, explaining what signals were detected, what the model assessed, and what (if anything) went wrong.
3. Worker can escalate to insurer review using this explanation as supporting context.

**Phase 2 — Human Reviewer Follow-Up:**

The Maintenance Check is simultaneously logged in the insurer's admin queue. A designated reviewer independently examines the same data — activity records, disruption signals, AI output, and SHAP breakdown — and sends a follow-up message to the worker confirming either:
- The AI explanation was accurate, or
- A correction has been made, or
- The issue has been escalated for a policy-level fix.

All messages are delivered in the worker's preferred language.

**Design constraints:**
- Maximum 3 uses per day per worker
- Does not automatically approve a claim — provides an auditable explanation
- Human reviewer follow-up targeted within a defined response window
- Available in all major Indian languages

---

## AI and ML Models

A core design principle: parametric triggers are threshold conditions that initiate a disruption log, but the AI layer sits above them to determine risk pricing, verify claim legitimacy, and forecast future exposure. The system is not a rule engine — thresholds are inputs to ML models, not the decision-makers.

### Model 1 — Dynamic Premium Calculation (XGBoost Regressor + SHAP)

**Purpose:** Predict expected weekly income loss probability per worker and zone.

**Input features:**
- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours on InDel
- Platform order density variance in the zone (InDel internal)
- Worker income stability score (earnings variance over past 8 weeks)

**Training data:** Synthetic dataset from IMD historical weather records, CPCB AQI archives, and simulated InDel order disruption logs. Continuous background retraining pipeline updates the model as real data accumulates.

**Output:** Continuous risk score (0–1) mapped to weekly premium.

**Retraining cadence:** Monthly in prototype; continuous pipeline in production.

**SHAP Explainability:**

Each worker's premium is explainable in terms of contributing factors:

```
Premium = Rs. 18 because:
  Flood risk in zone        +Rs. 6
  Rolling AQI pattern       +Rs. 3
  Income instability score  +Rs. 2
  Base rate                  Rs. 7
```

This breakdown is surfaced in the Maintenance Check feature and available to insurers for regulatory and audit purposes.

---

### Model 2 — Fraud Detection (Isolation Forest + DBSCAN + Rule Overlay)

See Step 7 above for full layer-by-layer description.

**Retraining cadence:** Weekly.

---

### Model 3 — Disruption Forecasting (Facebook Prophet — Time Series)

**Purpose:** Forward-looking forecast of likely disruption events and associated claim volume for the coming week, broken down by zone. Used exclusively for insurer reserve planning.

**Input:** Historical weather, AQI trends, InDel order volume history by zone and season.

**Output:** Zone-level claim probability for the next 7 days.

**Use:** Insurer reserve planning only — not used for individual claim decisions.

**Retraining cadence:** Weekly.

Prophet is selected for the prototype for its reliability on small datasets and strong handling of seasonal patterns.

---

### Model Card Summary

| Model | Type | Primary Input | Output | Retraining |
|---|---|---|---|---|
| Premium Calculator | XGBoost + SHAP | Zone risk features + worker profile | Weekly premium + SHAP breakdown | Monthly / continuous |
| Fraud Detector | Isolation Forest + DBSCAN + Rules | GPS + claim behavior + InDel activity | Anomaly score + cluster verdict + decision | Weekly |
| Disruption Forecaster | Prophet Time Series | Historical weather + InDel order logs | Zone claim probability (7-day) | Weekly |

---

### Future Model Upgrades

| Component | Prototype | Production |
|---|---|---|
| Risk Pricing | XGBoost + SHAP | Same, larger training corpus |
| Fraud Detection | Isolation Forest + DBSCAN + Rules | Same stack, larger corpus |
| Disruption Forecasting | Prophet | DeepAR → Temporal Fusion Transformer |

**DeepAR** — Probabilistic deep learning model that learns correlated disruption patterns across multiple zones simultaneously. Unlike Prophet (per-zone), DeepAR learns that a rainfall-driven order drop in Zone A predicts a similar drop in Zone B shortly after. Planned after 6+ months of real zone-level data.

**Temporal Fusion Transformer (TFT)** — Multi-horizon forecasting model combining attention mechanisms with recurrent networks. Outputs interpretable attention weights showing which inputs drove each forecast. Planned as a longer-term upgrade once compute infrastructure is in place.

---

## Weekly Premium Model

> All figures are illustrative estimates. Final amounts will be determined during development.

**Base structure:**
- Weekly premium range: Rs. 10 – Rs. 25 (dynamically calculated per worker per zone)
- Coverage ratio: Estimated 80–90% of calculated income loss
- Maximum weekly payout: Estimated Rs. 800
- Default: Automatic deduction from weekly earnings

**Sample zone premiums:**

| Worker Zone | Risk Level | Weekly Premium (est.) | Max Weekly Payout (est.) |
|---|---|---|---|
| Koramangala, Bengaluru | Low | Rs. 12 | Rs. 600 |
| Rohini, Delhi | Medium | Rs. 17 | Rs. 700 |
| Tambaram, Chennai | High | Rs. 22 | Rs. 800 |

**Payment options:**

- **Automatic deduction:** Deducted from weekly earnings at end of week.
- **Manual payment:** Worker pays at any point during the week, in full or split across days, as long as the total is settled before week-end.
- **Advance partial payment:** Worker pays approximately half the standard weekly premium as a lump sum (e.g., Rs. 200 if weekly premium is Rs. 20), covering the corresponding number of weeks plus two additional weeks at no charge. Normal weekly payments resume after this period.

**Non-payment consequences:**
- 1 missed week → coverage pauses from following week
- 2+ consecutive missed weeks → policy suspended
- Suspended policy requires re-enrollment and a fresh waiting period before claims are valid again

**Continuity rewards:**

Workers who maintain consistent payments without claims over time receive incremental benefits — reduced premiums, extended coverage periods, or increases to their maximum payout ceiling. Specific milestones to be defined during the product design phase.

---

## Fraud Prevention — Economic Activity Consistency Defense

To mitigate GPS spoofing and coordinated fraud, InDel shifts claim verification from location-based validation to **economic activity validation**.

**Core principle:** Instead of asking "Was the worker present in the disruption zone?", the system evaluates: "Did the worker experience a loss of earning opportunity consistent with other workers in the same zone?"

**Zone-Level Economic Impact Modeling:**

For each disruption event, the system constructs a baseline comparison — expected earnings pattern under normal conditions vs actual earnings during the disruption window — across all active workers in the affected zone.

**Worker-Level Consistency Evaluation:**

| Genuine Worker Characteristics | Fraudulent Behavior Indicators |
|---|---|
| Earnings drop consistent with peer workers | No meaningful change in activity pattern |
| Increased idle time due to reduced order availability | Lack of delivery attempts or unrealistic inactivity |
| Multiple failed or reduced delivery attempts | Earnings pattern deviates from zone-wide trend |

**Additional anti-spoofing signals:**
- Distance traveled from last confirmed delivery checkpoint (catches fake GPS positioning)
- Order pickup and non-pickup patterns
- Weather API cross-referencing with government alerts
- Group-level request scaling for high-volume disruption events

InDel verifies **economic impact**, not presence.

---

## Risk Controls and Edge Cases

### Edge Case 1 — Global Lockdown / Mass Correlated Disruption

**Problem:** Large-scale event hits all workers simultaneously. Premium pool risks depletion in one week.

**Response:**
- **Catastrophic Event Cap:** When aggregate claims exceed a defined percentage of the active pool in a single week, individual payouts are proportionally reduced. Formula: `Individual payout = Calculated entitlement × (Available pool / Total eligible claims)`
- **Reinsurance Layer:** Insurer purchases reinsurance activating when weekly aggregate claims exceed a set threshold. Included in the financial model; not in the hackathon prototype.
- **Lockdown Partial Coverage Clause:** Government-mandated full lockdowns are a special category. Coverage is capped at a reduced rate for up to 2 consecutive weeks; beyond that, coverage pauses and premiums are suspended. Disclosed at onboarding.

---

### Edge Case 2 — Zone Hopping (Deliberate Location Fraud)

**Problem:** Worker enrolls in a low-risk zone, then moves GPS to a high-risk zone before a disruption.

**Response:**
- **Zone Lock with Cooling Period:** When GPS detects a zone change, the new zone's risk profile immediately applies to premiums. A 7-day waiting period is enforced before claims in the new zone are eligible.
- **Mobility Pattern Scoring:** Zone-change frequency is a feature in the fraud model. Workers appearing in a high-risk zone with no prior activity history get a high anomaly score.
- **Premium Auto-Adjustment:** If GPS activity consistently shows the worker outside their declared zone over a rolling 2-week period, the risk profile is automatically reclassified.

---

### Edge Case 3 — Transit Disruption (Mid-Delivery Disruption)

**Problem:** Worker is mid-delivery when a disruption occurs between their start and destination. They may be offline. Their enrolled zone differs from the disruption location.

**Response:**

Transit Disruption Events are a distinct claim type. Coverage anchor is the active InDel delivery order, not the enrolled zone. Eligibility requires all four conditions to be satisfied:

1. Active InDel delivery order existed at the time of disruption
2. GPS trail shows directional movement consistent with the delivery route before stoppage
3. The disruption zone had a verified trigger active at the time of GPS stoppage
4. GPS stoppage occurred **after** the trigger fired, not before

If all four conditions are met, the claim is flagged as eligible. Zone-lock and home-zone rules do not apply to Transit Disruption Events. During mass events, the system falls back to zone-cluster verification rather than individual route tracing.

---

### Edge Case 4 — Interstate Travel

**Problem:** Worker's insurance is calibrated for their home state. Travel elsewhere puts the risk model outside its training data.

**Response:**
- **Under 72 hours:** Coverage travels with the worker using home zone parameters.
- **Over 72 hours:** System flags a zone migration event. Worker is prompted to update their registered zone. 7-day waiting period applies; premium recalculated at next weekly cycle.
- Interstate transit disruptions follow Transit Disruption Event logic — state boundaries do not affect coverage on an active InDel order.

---

### Edge Case 5 — Additional Cases

| Scenario | Detection Method | System Response |
|---|---|---|
| Global lockdown / mass event | Aggregate claims exceed pool threshold | Proportional reduction + reinsurance activation |
| Zone hopping | Mobility anomaly score + GPS zone mismatch | 7-day zone lock + premium auto-adjustment |
| Mid-delivery transit disruption | Active order + GPS trail + trigger timing | Eligible claim queued; worker files on reconnection |
| Interstate travel under 72 hours | GPS state detection | Home zone rules apply, coverage continues |
| Interstate travel over 72 hours | Persistent GPS state mismatch | Zone migration prompt + 7-day waiting period |
| Connectivity loss during disruption | Worker offline | Disruption logged; claim processed on reconnection |
| API failure | Internal signals only | System uses InDel order data as fallback |
| Overlapping disruptions | Temporal deduplication | Merged into single payout window |
| Worker joins after disruption | Onboarding timestamp check | Not eligible |
| Inactive worker during event | Activity check | No payout |
| Event but no income loss | Loss calculation check | No payout |
| Multiple IDs / accounts | Identity linking | Accounts merged, payout capped at single-worker limit |

---

## Illustrative Unit Economics

> Conservative assumptions for a cohort of 1,000 active workers in Chennai during a standard month. These figures validate directional viability, not final projections.

| Assumption | Value |
|---|---|
| Average weekly premium | Rs. 17 |
| Active weeks per month | 4 |
| Disruption events per worker per month | 0.08 |
| Average payout per event | Rs. 550 |

| Metric | Value |
|---|---|
| Total premium collected | Rs. 68,000 |
| Expected total payouts | Rs. 44,000 |
| Gross margin before ops cost | Rs. 24,000 (35%) |
| Projected loss ratio | ~65% |

A 65% loss ratio is within the acceptable range for microinsurance. Standard health microinsurance in India operates at 70–85%.

| City | Risk Profile | Avg. Weekly Premium | Expected Monthly Loss Ratio |
|---|---|---|---|
| Chennai | High (monsoon + heat) | Rs. 22 | 72% |
| Bengaluru | Medium | Rs. 16 | 61% |
| Pune | Low | Rs. 11 | 54% |

---

## Scenario Walkthroughs

**Scenario 1 — Flood Event (Chennai, August)**
Worker in Tambaram, earning Rs. 4,200/week, premium Rs. 22. Heavy rainfall logged above threshold at 11:40 AM. System logs disruption. Worker files claim at 1:00 PM. GPS confirms zone presence, InDel confirms order drop, fraud check passes. Payout of ~Rs. 360 approved via UPI.

**Scenario 2 — Heat Wave (Delhi, May)**
Worker in Rohini, earning Rs. 3,800/week, premium Rs. 19. Temperature threshold crossed at 1:00 PM. Disruption logged. InDel shows significant order drop. GPS in zone confirmed. Payout for 4-hour window (~Rs. 270) approved.

**Scenario 3 — Continuity Reward (Pune, February)**
Worker in Kothrud, earning Rs. 3,500/week, premium Rs. 11. Several consecutive weeks without a claim. System applies continuity reward — reduced premium or extended coverage period. Worker notified.

**Scenario 4 — Transit Disruption (Mid-Delivery Flood)**
Worker doing an active delivery from Adyar to Velachery, Chennai. Flash flood logged in Guindy (between the two points) at 3:12 PM. InDel confirms active order. GPS shows rider stopped in Guindy at 3:18 PM, after the trigger fired. All four transit conditions satisfied. Worker is offline — claim queued. On reconnection, worker is notified and files. Payout processed.

**Scenario 5 — Zone Hopping (Fraud Caught)**
Worker based in Pune (low risk, premium Rs. 11) moves GPS to Chennai flood zone the day before a major rainfall event. Mobility model detects zone shift with no Chennai activity history. Anomaly score high. Zone lock active — Chennai claims ineligible for 7 days. Filed claim auto-rejected. Premium auto-adjusts to Chennai risk profile from next cycle.

**Scenario 6 — Maintenance Check (Worker Disputes Eligibility)**
Worker believes they should have been eligible for a claim. Triggers Maintenance Check from dashboard. System calls AI API with activity data, zone disruption signals, and SHAP breakdown. Response returned in Tamil explaining what was detected and why eligibility did not trigger. Simultaneously, check is logged in insurer admin queue. Human reviewer examines the same data and sends a follow-up in Tamil — confirming accuracy, noting a correction, or flagging a model issue for escalation.

**Scenario 7 — National Lockdown**
78% of InDel workers across all zones file claims in one week. Aggregate claims exceed pool threshold. Catastrophic Cap activated. Individual payouts reduced proportionally. Workers notified. Reinsurance layer activated for insurer. From week 3: Lockdown Partial Coverage Clause applies — reduced payout rate, premiums suspended.

---

## Dashboards

### Worker Dashboard
- Active coverage status and current weekly premium
- Earnings this week vs estimated protected income
- Disruption alerts active in their zone
- Claim history and wallet balance
- Continuity reward progress
- Maintenance Check (up to 3/day) — shows AI response and human reviewer follow-up when available
- Language preference setting

### Platform Admin Dashboard
- Live order allocation volume by zone
- Active worker sessions and GPS distribution
- Disruption alerts and affected zone overlay
- Delivery completion rates and average order time

### Insurer Dashboard
- Premium pool health: collected vs paid out this week
- Loss ratio by zone and city
- Active claims in processing pipeline
- Fraud-flagged claims queue
- Forecasted claim volume for next 7 days (Prophet output)
- Reserve recommendation based on disruption forecast
- Maintenance Check review queue with full AI output and worker data visible per check

---

## Compliance and Regulatory Considerations

**Product Classification:** Parametric income protection falls under general insurance. The deploying insurer would file with IRDAI as a group microinsurance policy — a simplified approval pathway.

**Data Privacy:** Worker data falls under the Digital Personal Data Protection Act 2023. Architecture separates PII from risk modelling inputs. Raw GPS trails are not stored beyond the claim verification window (72 hours post-disruption).

**Consent:** Insurance enrollment is opt-in. Workers can pause or cancel coverage at any time. Premium deductions require explicit consent at enrollment.

**Payout Classification:** Payouts are compensation for income loss, not indemnity for an insured asset. Payouts below Rs. 2,50,000 annually are unlikely to create tax obligations for gig workers at current income levels.

**Language Support:** All worker-facing communications — including Maintenance Check outputs — will support all major Indian languages via a translation layer. This is a core accessibility requirement, not an optional feature.

> Note: A production deployment would require the deploying insurer to handle IRDAI product registration and KYC/AML obligations. These are outside the scope of the hackathon prototype.

---

## Tech Stack

| Layer | Planned Technology |
|---|---|
| Backend | Python (FastAPI) |
| Frontend | React.js |
| Database | PostgreSQL |
| Message Queue (Async Payouts) | Kafka / RabbitMQ |
| AI / ML (Prototype) | scikit-learn, XGBoost, SHAP, Prophet, DBSCAN, Isolation Forest |
| AI / ML (Future) | DeepAR, Temporal Fusion Transformer |
| Weather API | OpenWeatherMap (free tier) |
| AQI API | OpenAQ / WAQI |
| Traffic / Zone Alerts | Mock API (simulated) |
| Payment | Razorpay test mode / UPI simulator |
| Hosting | AWS / Render |

---

## Team ImaginAI

| Name | Role |
|---|---|
| Shravanthi Satyanarayanan | Backend & AI/ML |
| Gayathri U | Frontend & UX |
| Rithanya K A | Insurance Model & Research |
| Saravana Priyaa | Delivery Platform & DevOps |
| Subikha MV | System Design & Integration |

---

*Submitted for Guidewire DEVTrails 2026 — University Hackathon*
