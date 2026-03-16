# InDel - Insure,Deliver 
# Integrated Delivery and Income Protection Platform

**Team:** ImaginAI
**Hackathon:** Guidewire DEVTrails 2026
**Persona:** Food Delivery Partners (Swiggy / Zomato style)
**Current Phase:** Phase 1 — Ideation and Foundation

---

## The Problem

India's gig delivery workers earn based on completed orders. When external disruptions occur — heavy rain, extreme heat, severe pollution, curfews, or sudden order drops — deliveries stop and income disappears immediately. Workers lose 20–30% of monthly earnings during these events with no financial safety net.

Traditional insurance does not address this. It covers accidents, vehicles, and health. It does not cover the most real risk a delivery worker faces: losing a day's wages because the world made it impossible to work.

Existing parametric insurance solutions that attempt to solve this face a structural problem: they depend on third-party delivery platforms to share worker activity data. Platforms have no incentive to share this data, making verification unreliable and fraud detection weak.

---

## What We Plan to Build

InDel (Income Defense for Delivery Workers) will be an **AI-powered delivery ecosystem with parametric income protection built directly into the platform**.

InDel will not be an insurance product bolted onto an existing delivery app, nor a standalone insurance app requiring workers to trust a new product. The plan is a single integrated system where the delivery platform and the insurance engine share the same data layer from the ground up.

This integration means InDel will own the worker activity data, order volume data, GPS trails, and earnings history — the exact data that parametric income insurance requires to function accurately. There will be no dependency on third-party API access and no data gap between what the insurer needs and what the platform can provide.

```
Traditional Approach:
Swiggy/Zomato → API (if available) → Insurance Layer → Incomplete data → Weak fraud detection

InDel Approach:
Single Platform → Delivery System + Insurance Engine → Complete data → Accurate payouts + Strong fraud detection
```

---

## The Three Stakeholders

**Insurance Provider (Primary B2B Customer)**
Will deploy InDel as a white-label product vertical. Gains access to a previously uninsured, high-volume customer segment with a fully integrated data pipeline, automated claims processing, and actuarially sound risk modelling — without needing to negotiate data-sharing agreements with delivery platforms.

**Delivery Worker (End Beneficiary)**
Will use InDel as their delivery platform. Insurance coverage will be automatic and embedded — not a separate product to manage. Workers will receive compensation payouts without filing anything.

**Platform Administrator**
Will manage delivery operations, worker assignments, and zone management through a unified admin dashboard with full visibility into the insurance engine's performance.

---

## Proposed System Architecture

InDel will be composed of four tightly integrated layers sharing a single data backbone.

```
+----------------------------------------------------------+
|                     InDel Platform                       |
|                                                          |
|  +------------------+     +-------------------------+   |
|  |  Delivery Engine |     |    Insurance Engine     |   |
|  |                  |     |                         |   |
|  | Order Management |<--->| Policy Management       |   |
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

## Planned Platform Flow

```
Customer Places Order
        |
        v
Platform Assigns Delivery Task to InDel Worker
        |
        v
Worker Activity Continuously Recorded (GPS, session, orders)
        |
        v
AI Engine Monitors Environment + Internal Order Patterns
        |
        v
Disruption Detected (weather / AQI / order drop / zone closure)
        |
        v
Income Loss Calculated from Worker Earnings Baseline
        |
        v
Fraud Verification (GPS + activity + anomaly model)
        |
        v
Claim Auto-Approved
        |
        v
Instant Payout to Worker via UPI / Wallet
```

---

## How the System Is Designed to Work

### Step 1 — Worker Onboarding

Workers will register on the InDel platform as delivery partners. Onboarding will collect:

- Name, location, working zone
- Preferred working hours
- Bank account or UPI ID for payouts
- Delivery vehicle type

Insurance enrollment will be automatic at onboarding — confirmed with a single action. There will be no separate insurance application and no separate KYC. Delivery registration and insurance enrollment will happen in one flow. The system will immediately build an initial risk profile using the worker's declared zone and historical disruption data for that zone.

---

### Step 2 — Delivery Operations

Workers will receive and complete delivery assignments through the InDel platform. The system will continuously record:

- Active session timestamps
- GPS location and movement trails
- Orders assigned, accepted, completed, and dropped
- Earnings per order and per session
- Zone activity patterns over time

This data will serve dual purpose: powering delivery operations and continuously refining the worker's risk profile and income baseline for insurance calculations.

---

### Step 3 — AI Risk Profiling and Weekly Premium Calculation

At the start of each week, the AI engine will recalculate the worker's premium based on their updated risk profile.

The premium model is planned to use an XGBoost Regressor trained on:

- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history by city)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours over the past 4 weeks on InDel
- Platform-level order density variance in the zone (InDel internal data)
- Worker's income stability score (variance in weekly earnings over past 8 weeks)

Training data will be a synthetic dataset generated from IMD historical weather records, CPCB AQI archives, and simulated InDel platform disruption and order logs.

The model will output a continuous risk score between 0 and 1. This score will be applied against a base premium and coverage multiplier to produce the final weekly premium. Two workers in the same city but different zones will receive meaningfully different premiums based on learned zone-level risk — not a fixed city-wide rate.

**Planned premium examples:**

| Worker Zone | Risk Level | Weekly Premium | Max Weekly Payout |
|---|---|---|---|
| Koramangala, Bengaluru (low flood risk) | Low | Rs. 12 | Rs. 600 |
| Rohini, Delhi (heat zone) | Medium | Rs. 17 | Rs. 700 |
| Tambaram, Chennai (flood-prone) | High | Rs. 22 | Rs. 800 |

Premiums will be deducted automatically from the worker's weekly earnings within the platform — no separate payment step required.

---

### Step 4 — Parametric Trigger Monitoring

The system will continuously poll external data sources and internal platform metrics. When a trigger threshold is crossed, a claim evaluation will be automatically initiated — no worker action required.

**Planned triggers:**

| Disruption Type | Trigger Condition | Data Source |
|---|---|---|
| Heavy Rain | Rainfall > 35mm in 3 hours in worker's zone | OpenWeatherMap |
| Extreme Heat | Temperature > 43 degrees C during active hours | OpenWeatherMap |
| Severe Pollution | AQI > 300 (Hazardous) in worker's city | OpenAQ / WAQI |
| Curfew / Bandh | Verified zone closure | Traffic API / Government alert |
| Platform Order Drop | Order volume drop > 80% for > 2 hours in zone | InDel internal data |
| Flash Flood | Flood alert issued for zone by IMD | Weather alert API |

Because InDel will own its platform data, the order volume drop trigger will be particularly powerful — no external API needed. The system will detect a collapse in delivery activity across a zone in real time using its own order management data.

---

### Step 5 — Income Loss Calculation

When a trigger fires, the system will calculate the worker's estimated income loss for the disruption period.

```
Baseline hourly rate     = Average hourly earnings over past 4 weeks (from InDel data)
Disruption window        = Time from trigger start to trigger end
Expected earnings        = Baseline hourly rate x disruption hours
Actual earnings          = Earnings recorded in InDel during disruption window
Income loss              = Expected earnings - Actual earnings
Payout amount            = Income loss x coverage ratio (capped at weekly maximum)
```

**Illustrative example:**

A worker earns an average of Rs. 120/hour over 4 weeks. A flood trigger fires at 11:40 AM and clears at 5:30 PM (5 hours 50 minutes).

```
Expected earnings:  Rs. 120 x 5.83 hrs  = Rs. 700
Actual earnings:    2 partial deliveries = Rs. 80
Income loss:        Rs. 700 - Rs. 80     = Rs. 620
Payout (90% ratio): Rs. 558
```

---

### Step 6 — Automated Claim Processing

Once income loss is calculated, the fraud detection layer will run verification before payout approval.

Planned standard verification (target under 30 seconds):

- Was the worker logged into InDel during the disruption window?
- Does GPS confirm presence in the affected zone?
- Does InDel order activity confirm reduced earnings during the window?
- Is the claim behavior consistent with the zone-wide claim cluster?

If all checks pass, the claim will be auto-approved and payout initiated. The target pipeline from trigger detection to payout initiation is under 15 minutes for standard claims.

---

### Step 7 — Instant Payout

Approved payouts will be sent to the worker via:

- UPI direct transfer
- InDel in-app wallet (usable against future premium deductions)
- Bank transfer (next-day settlement for amounts above Rs. 500)

Workers will receive a push notification explaining the disruption detected, the income loss calculated, and the amount paid. This transparency is a deliberate design choice to build long-term trust in the system.

---

## Weekly Premium Model

**Planned base structure:**
- Weekly premium range: Rs. 10 — Rs. 25 (dynamically calculated per worker per zone)
- Coverage ratio: 80–90% of calculated income loss
- Maximum weekly payout: Rs. 800
- Premium deduction: Automatic from weekly InDel platform earnings

**Loyalty mechanics:**

To address the psychological barrier of paying premiums without seeing claims, we plan to include a retention layer:

| Milestone | Reward |
|---|---|
| 8 consecutive weeks without a claim | Rs. 50 wallet credit |
| 12 consecutive weeks without a claim | One week premium waived, coverage continues |
| 6+ months active | Maximum payout increased by 10% |

This will maintain premium flow for the insurer while giving workers a tangible reason to stay enrolled.

---

## Planned AI and ML Integration

A core design principle of InDel is that parametric triggers will be threshold conditions that initiate a claim event, but the AI layer will sit above them to determine risk pricing, verify legitimacy, and predict future exposure. The system is not intended to be a rule engine. Thresholds are inputs to ML models, not the decision-makers themselves.

---

### Model 1 — Dynamic Premium Calculation (XGBoost Regressor)

Will predict the expected weekly income loss probability for a given worker profile and zone. Output feeds directly into premium pricing.

**Planned input features:**
- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours on InDel
- Platform order density variance in the zone (InDel internal)
- Worker income stability score (earnings variance over past 8 weeks)

**Training data:** Synthetic dataset from IMD historical weather records, CPCB AQI archives, and simulated InDel order disruption logs.

**Output:** Continuous risk score (0–1) mapped to weekly premium.

**Planned retraining cadence:** Monthly.

---

### Model 2 — Fraud Detection (Isolation Forest + Rule Overlay)

**Layer 1 — Isolation Forest anomaly detection:**

Will be trained on expected claim behavior patterns across the worker pool. Flags workers whose claim profile deviates statistically from the zone-wide cluster.

Planned input features:
- GPS trail consistency during disruption window
- Ratio of claimed loss to historical earnings baseline
- Claim frequency per worker over rolling 8-week window
- Zone-wide claim clustering (are other workers in the same zone also claiming?)
- Mobility pattern score (stability of worker's operating zone week over week)

**Layer 2 — Rule overlay for hard disqualifiers:**
- Worker GPS not in the affected zone at trigger time: auto-reject
- InDel platform shows completed deliveries during claimed disruption window: auto-reject
- Anomaly score above 0.85: route to manual review queue

The ML layer will catch soft anomalies that rules would miss. The rule layer will handle clear disqualifiers deterministically without unnecessary compute.

---

### Model 3 — Disruption Forecasting (Facebook Prophet — Time Series)

Will be a forward-looking model forecasting likely claim events for the coming week, broken down by zone. Will feed the insurer dashboard for reserve planning ahead of high-risk periods.

**Planned input:** Historical weather, AQI trends, InDel order volume history by zone and season.

**Output:** Zone-level claim probability for next 7 days.

**Use:** Insurer reserve planning only — not individual claim decisions.

**Planned retraining cadence:** Weekly.

---

### Model Card Summary

| Model | Type | Primary Input | Output | Retraining |
|---|---|---|---|---|
| Premium Calculator | XGBoost Regressor | Zone risk features + worker profile | Weekly premium (Rs.) | Monthly |
| Fraud Detector | Isolation Forest + Rules | GPS + claim behavior + InDel activity | Anomaly score + decision | Weekly |
| Disruption Forecaster | Prophet Time Series | Historical weather + InDel order logs | Zone claim probability | Weekly |

---

## Illustrative Unit Economics

The following uses conservative assumptions for a cohort of 1,000 active workers in Chennai during a standard month, to validate that the financial model is viable.

**Assumptions:**
- Average weekly premium: Rs. 17
- Active weeks per month: 4
- Disruption events per worker per month: 0.8 (roughly one event every 5–6 weeks, based on Chennai IMD historical data)
- Average payout per event: Rs. 550

**Monthly figures:**

| Metric | Value |
|---|---|
| Total premium collected | Rs. 68,000 |
| Expected total payouts | Rs. 44,000 |
| Gross margin before ops cost | Rs. 24,000 (35%) |
| Projected loss ratio | ~65% |

A 65% loss ratio is within acceptable range for microinsurance products. Standard health microinsurance in India operates at 70–85%. InDel's parametric structure should keep loss ratios predictable because payouts will be capped and calculated algorithmically — no inflated settlements.

**City-level comparison:**

| City | Risk Profile | Avg. Weekly Premium | Expected Monthly Loss Ratio |
|---|---|---|---|
| Chennai | High (monsoon + heat) | Rs. 22 | 72% |
| Bengaluru | Medium | Rs. 16 | 61% |
| Pune | Low | Rs. 11 | 54% |

---

## Scenario Walkthroughs

These scenarios illustrate how the system is designed to behave across different disruption types. They are based on our current design assumptions and will be validated through simulation during development.

**Scenario 1 — Flood Event (Chennai, August)**
Worker: InDel rider, Tambaram, earns Rs. 4,200/week, premium Rs. 22.
Event: 48mm rainfall in 3 hours. Trigger fires at 11:40 AM.
Planned response: Weather API flags trigger. InDel confirms 91% order drop in zone. GPS confirms worker present. Income loss calculated: Rs. 360. Fraud check passes. Payout initiated at 11:52 AM. Worker receives Rs. 360 via UPI at 11:54 AM.
Target time from trigger to payout: 14 minutes.

**Scenario 2 — Heat Wave (Delhi, May)**
Worker: InDel rider, Rohini, earns Rs. 3,800/week, premium Rs. 19.
Event: Temperature reaches 45 degrees C at 1:00 PM.
Planned response: Temperature API flags trigger. InDel shows 74% order drop. GPS in zone confirmed. Payout Rs. 270 for 4-hour window initiated automatically.

**Scenario 3 — No Claim Loyalty Reward (Pune, February)**
Worker: InDel rider, Kothrud, earns Rs. 3,500/week, premium Rs. 11.
Event: 8 consecutive weeks without a claim.
Planned response: Rs. 50 wallet credit applied automatically. Week 9 premium waived. Coverage continues uninterrupted.

**Scenario 4 — Transit Disruption (Mid-Delivery Flood)**
Worker: InDel rider, active delivery from Adyar to Velachery, Chennai. Premium Rs. 20.
Event: Flash flood in Guindy (between Adyar and Velachery) at 3:15 PM.
Planned response: InDel confirms active order at 3:15 PM. GPS shows rider stopped in Guindy at 3:18 PM. Flood trigger confirmed in Guindy at 3:12 PM. All four transit verification conditions satisfied. Claim auto-approved without worker action. Worker offline. Payout Rs. 180 queued. Delivered at 5:40 PM on reconnection.

**Scenario 5 — Zone Hopping Attempt (Fraud Caught)**
Worker: InDel rider, Pune (low risk), premium Rs. 11.
Event: Worker relocates GPS to Chennai flood zone the day before a major rainfall event.
Planned response: Mobility model detects 1,400km zone shift with no Chennai activity history. Anomaly score: 0.94. Zone lock active — Chennai claims ineligible for 7 days. Claim rejected. Premium auto-adjusts to Rs. 22 from next weekly cycle.

**Scenario 6 — National Lockdown**
Event: Government announces national lockdown. 78% of InDel workers across all zones trigger claims in one week.
Planned response: Aggregate claims exceed 55% of pool threshold. Catastrophic Cap activated. Individual payouts reduced to 58% of entitlement. Workers notified via app. Reinsurance layer activated for insurer. From week 3: Lockdown Clause applies — 50% payout, premiums suspended.

---

## Risk Controls and Edge Cases

This section documents how InDel is designed to handle failure modes and adversarial scenarios. Addressing these edge cases early is a core part of our design philosophy — we want to build a system that is honest about its limits from the start.

---

### Edge Case 1 — Global Lockdown or Mass Correlated Disruption

**The problem:** A pandemic lockdown or city-wide catastrophe hits every worker simultaneously. The premium pool risks depletion in a single week. This is correlated risk — the primary actuarial failure mode for parametric insurance at scale.

**Our design approach:**

A Catastrophic Event Cap will activate when aggregate claims exceed 55% of the active premium pool in a single week. Individual payouts will be proportionally reduced. Workers will receive reduced payouts, not zero.

Planned formula: Individual payout = Calculated entitlement x (Available pool / Total eligible claims)

A Reinsurance Layer is modelled into the insurer deployment architecture. The deploying insurer would purchase reinsurance activating when weekly aggregate claims exceed 60% of the collected pool. This is not part of the hackathon prototype but is explicitly included in the financial model to demonstrate production viability.

A Lockdown Partial Coverage Clause will define government-mandated full lockdowns as a special disruption category. Coverage will be capped at 50% of normal payout for up to 2 consecutive weeks. Beyond 2 weeks, coverage pauses and premiums are suspended. This will be disclosed to workers at onboarding.

---

### Edge Case 2 — Zone Hopping (Deliberate Location Fraud)

**The problem:** A worker enrolls in a low-risk zone at a cheaper premium, then physically relocates to a high-risk zone before a disruption to claim a payout they underpaid for.

**Our design approach:**

Zone Lock with Cooling Period: When GPS detects a worker's active zone has changed, the new zone's risk profile will immediately apply to premium calculation. A 7-day waiting period will be enforced before claims in the new zone are eligible.

Mobility Pattern Scoring: The fraud model will include zone-change frequency as a feature. A worker with a stable operating radius for months who suddenly appears in a flood zone on trigger day will receive a high anomaly score and be routed to manual review.

Premium Auto-Adjustment: If GPS activity consistently shows the worker outside their declared zone over a rolling 2-week period, the system will reclassify their risk profile to reflect actual operating location.

---

### Edge Case 3 — Transit Disruption (Disruption Between Delivery Points)

**The problem:** A worker is mid-delivery from A to B when a disruption occurs at C between them. The delivery stalls. The worker has no connectivity to file anything. Their enrolled zone differs from C.

**Our design approach:**

Transit Disruption Events will be a distinct claim type where the enrolled zone is irrelevant. The coverage anchor will be the active InDel delivery order.

Automatic verification will use four conditions:
- Active InDel delivery order existed at the time of disruption
- GPS trail shows directional movement consistent with the delivery route before stoppage
- Point C had a verified disruption trigger active at the time of GPS stoppage
- GPS stoppage occurred after the trigger fired in C, not before

If all four conditions are met, the claim will be auto-approved by the system. The payout will be queued and delivered on connectivity restoration. Zone-lock and home-zone rules will not apply to Transit Disruption Events.

Scalability consideration: Individual GPS route verification is computationally expensive at high volume. During mass disruption events, the system will fall back to zone-cluster verification — was the worker's GPS in the disrupted zone, did they have an active InDel session, did platform-wide order completion rates drop in that zone. Individual route tracing will be reserved for anomaly-flagged claims only, keeping compute proportional to actual fraud risk.

---

### Edge Case 4 — Interstate Travel

**The problem:** A worker's insurance is priced for their home state. Travel to another state puts the risk model outside its training data. The insurer may not hold product licenses in all states.

**Our design approach:**

Home Zone Anchor with Portable Coverage: The policy will be anchored to the registered state at enrollment. Coverage will travel with the worker for up to 72 hours in another state using home zone risk parameters — mirroring how vehicle insurance operates for interstate travel in India.

Zone Migration for Extended Stays: If GPS shows the worker consistently in a new state beyond 72 hours, a zone migration event will be flagged. The worker will be prompted to update their registered zone. A 7-day waiting period will apply before claims in the new zone are valid. Premium will be recalculated for the new state's risk profile at the start of the next weekly cycle.

Regulatory Handling: Interstate coverage portability will be managed at the insurer level through a group microinsurance product structure filed with IRDAI, allowing nationwide coverage under a single product registration.

Interstate Transit Disruptions will follow the Transit Disruption Event logic above. State boundaries will not affect coverage eligibility when the worker is mid-delivery on an active InDel order.

---

### Edge Case Summary

| Scenario | Detection Method | Planned System Response |
|---|---|---|
| Global lockdown / mass event | Aggregate claims exceed 55% of pool | Proportional payout reduction + reinsurance activation |
| Zone hopping | Mobility anomaly score + GPS zone mismatch | 7-day zone lock + premium auto-adjustment |
| Mid-delivery transit disruption | Active order + GPS trail + trigger timing | Auto-approved Transit Disruption Event, no worker action |
| Interstate travel under 72 hours | GPS state detection | Home zone rules apply, coverage continues |
| Interstate travel over 72 hours | Persistent GPS state mismatch | Zone migration prompt + 7-day waiting period |
| Connectivity loss during disruption | Worker cannot file | System files automatically, payout queued on reconnection |

---

## Planned Dashboards

### Worker Dashboard
- Active coverage status and current weekly premium
- Earnings this week vs protected income baseline
- Active disruption alerts in their zone
- Payout history and wallet balance
- No-claim bonus progress tracker

### Platform Admin Dashboard
- Live order volume by zone
- Active worker sessions and GPS distribution map
- Disruption alerts and affected zone overlay
- Delivery completion rates and average order time

### Insurer Dashboard
- Premium pool health: collected vs paid out this week
- Loss ratio by zone and city
- Active claims in processing pipeline
- Fraud-flagged claims queue
- Prophet model output: predicted claim volume for next 7 days
- Reserve recommendation based on forecasted disruption risk

---

## Compliance and Regulatory Considerations

**Product Classification:** Parametric income protection falls under general insurance. The deploying insurer would file with IRDAI as a group microinsurance policy, which carries a simplified approval pathway compared to individual policies.

**Data Privacy:** Worker data collected through InDel will fall under the Digital Personal Data Protection Act 2023. The planned architecture separates PII from risk modelling inputs and will not store raw GPS trails beyond the claim verification window (72 hours post-disruption).

**Consent:** Insurance enrollment will be opt-in and explicitly confirmed at onboarding. Workers will be able to pause or cancel coverage at any time. Premium deductions will require active consent.

**Payout Classification:** Parametric payouts will be compensation for income loss, not indemnity for an insured asset. Payouts below Rs. 2,50,000 annually are unlikely to create tax obligations for gig workers at current income levels.

Note: A production deployment would require the deploying insurer to handle IRDAI product registration and KYC/AML obligations. These are out of scope for the hackathon prototype.

---

## Tech Stack

| Layer | Planned Technology |
|---|---|
| Backend | Python (FastAPI) |
| Frontend | React.js |
| Database | PostgreSQL |
| AI / ML | scikit-learn, XGBoost, Prophet |
| Weather API | OpenWeatherMap (free tier) |
| AQI API | OpenAQ / WAQI |
| Traffic / Zone Alerts | Mock API (simulated) |
| Payment | Razorpay test mode / UPI simulator |
| Hosting | AWS / Render |

---

## Development Roadmap

**Phase 1 (Current — Due March 20)**
- This README and idea documentation
- Initial user research with gig workers
- System architecture diagram
- Financial model validation

**Phase 2 (March 21 — April 4)**
- Worker registration and onboarding flow
- Basic delivery order management
- Dynamic premium calculation engine
- Parametric trigger logic (5 triggers)
- Claims management module

**Phase 3 (April 5 — April 17)**
- Advanced fraud detection layer
- Transit Disruption Event handler
- Simulated instant payout integration
- Three-dashboard system: Worker, Admin, Insurer
- Final pitch deck and 5-minute demo video

---

## Why This Approach

Most teams at this hackathon will build a standalone insurance app that depends on third-party delivery platform APIs they cannot realistically access. InDel's approach eliminates this dependency by building the delivery platform and the insurance engine as one system.

This means InDel's fraud detection will be grounded in real first-party activity data rather than self-reported claims. The risk model will be trained on actual order patterns rather than approximations. The income loss calculation will use verified earnings history rather than declared averages.

The parametric model is designed to eliminate the single biggest operational cost in microinsurance: claims processing. Traditional microinsurance claims take 3–7 days and require human review. Our target pipeline of under 15 minutes from trigger to payout with zero human involvement for standard claims represents an estimated 80–90% reduction in claims processing overhead for the deploying insurer.

The integrated platform addresses the single biggest adoption barrier: distribution. No separate app, no marketing spend per worker, no trust gap — insurance embedded into the platform the worker already uses every day.

---

## Team ImaginAI

| Name | Role |
|---|---|
| Shravanthi Satyanarayanan | Backend & AI/ML |
| Gayathri U | Frontend & UX |
| Rithanya K | Insurance Model & Research |
| Saravana Priyaa | Delivery Platform & DevOps |
| Subikha MV | System Design & Integration |

---

*Submitted for Guidewire DEVTrails 2026 — University Hackathon*
