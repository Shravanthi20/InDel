# InDel — Insure, Deliver
A platform that combines **delivery management and parametric income insurance** so gig workers are protected when disruptions stop them from working.

**Team:** ImaginAI
**Hackathon:** Guidewire DEVTrails 2026
**Persona:** E-commerce 
**Current Phase:** Phase 1 (Ideation)

> **Note:** This README describes what we are planning to build. Specific values such as premium amounts, payout figures, coverage ratios, and threshold numbers are illustrative estimates used for design and modelling purposes. As stated in the hackathon guidelines, exact figures do not need to be defined at this stage and will be refined during development in collaboration with the relevant stakeholders. Any API integrations or third-party service references are subject to change as the project progresses.

---

## The Problem

India's gig delivery workers earn based on completed orders. When external disruptions occur — heavy rain, extreme heat, severe pollution, curfews, or sudden order drops — deliveries stop and income drops significantly. Workers can lose 20–30% of monthly earnings during these events with no financial protection in place.

Traditional insurance covers accidents, vehicles, and health. Income lost because conditions made it impossible to work falls outside the scope of any existing product available to these workers.

Existing parametric insurance solutions that attempt to address this face a structural problem: they depend on third-party delivery platforms to share worker activity data. Platforms have no strong incentive to share this data, making verification unreliable and fraud detection weak.

---

## What We Plan to Build

InDel (Income Defense for Delivery Workers) is planned as a **B2B platform built for insurance providers**, combining a delivery management system and a parametric income protection engine into a single integrated product.

The insurance provider is the primary customer. InDel gives them a ready-to-deploy infrastructure that handles delivery worker management, real-time disruption monitoring, income loss calculation, claim verification, and payout processing — all in one system, without depending on data from third-party delivery apps.

On the delivery side, InDel will handle delivery allocation and assignment for partner workers — similar to how delivery management platforms work, not as a consumer-facing food ordering app. Workers log in, receive assignments, complete deliveries, and earn. The insurance layer runs in the background using the same activity data the delivery system already collects.

This is the core advantage: because InDel owns both sides of the data, the insurance engine has accurate, verifiable, first-party information to work with at all times.

```
Traditional Approach:
Insurance Provider → needs data from Swiggy/Zomato → API access unlikely → incomplete data → weak fraud detection

InDel Approach:
Insurance Provider deploys InDel → delivery system + insurance engine share one data layer → accurate verification → reliable payouts
```

---

## Stakeholders

**Insurance Provider (Primary B2B Customer)**
Purchases and deploys InDel. Gets access to a previously uninsured worker segment with an integrated data pipeline, automated claim processing, and risk analytics — without needing to negotiate data agreements with existing delivery platforms.

**Delivery Platform Partner (e.g. Swiggy, Zomato)**
InDel is not replacing consumer delivery apps. The collaboration model is that InDel handles the delivery-side worker management and insurance layer, while the platform partner benefits from having their delivery workers covered and financially protected. The platform's role here is delivery allocation and worker coordination — InDel sits alongside that and handles the protection layer.

**Delivery Worker (End Beneficiary)**
Uses InDel for delivery assignments. Can choose to opt into the income protection plan at onboarding or at any point after. If enrolled, their coverage runs in the background based on their actual activity — they do not need to manage it actively.

---

## Proposed System Architecture

InDel will be composed of four integrated layers sharing a single data backbone.

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

## Planned Platform Flow

```
Worker Receives Delivery Assignment via InDel
        |
        v
Worker Activity Recorded (GPS, session, orders, earnings)
        |
        v
AI Engine Monitors Environment + Internal Order Patterns
        |
        v
Disruption Detected (weather / AQI / order drop / zone closure)
        |
        v
Income Loss Estimated from Worker Earnings Baseline
        |
        v
Worker Files Claim Request (or uses Maintenance Check)
        |
        v
Fraud Verification (GPS + activity + anomaly model)
        |
        v
Claim Approved
        |
        v
Payout to Worker via UPI / Wallet
```

---

## How the System Is Designed to Work

### Step 1 — Worker Onboarding

Workers will register on InDel as delivery partners. Onboarding will collect:

- Name, location, working zone
- Preferred working hours
- Bank account or UPI ID for payouts
- Delivery vehicle type

Income protection enrollment is **optional** and presented as a separate choice during onboarding. Workers who decline can still use the platform for deliveries and can choose to enroll in the protection plan at any later point from their dashboard. Workers who enroll will have their coverage start from the following weekly cycle.

---

### Step 2 — Delivery Operations

Workers will receive and complete delivery assignments through InDel. The system will continuously record:

- Active session timestamps
- GPS location and movement trails
- Orders assigned, accepted, completed, and dropped
- Earnings per order and per session
- Zone activity patterns over time

This data serves dual purpose: powering delivery allocation and continuously updating the worker's risk profile and earnings baseline for insurance calculations.

---

### Step 3 — AI Risk Profiling and Weekly Premium Calculation

At the start of each week, the AI engine will calculate the worker's premium based on their current risk profile.

The premium model is planned to use an XGBoost Regressor trained on:

- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history by city)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours over the past 4 weeks
- Platform-level order density variance in the zone (InDel internal)
- Worker's income stability score (variance in weekly earnings over past 8 weeks)

Training data in the early phase will be a synthetic dataset generated from IMD historical weather records, CPCB AQI archives, and simulated InDel platform disruption and order logs. As the platform accumulates real delivery data, the model will be updated through a continuous retraining pipeline — this runs in the background without requiring any manual intervention.

The model will output a continuous risk score mapped to a weekly premium. Two workers in the same city but different zones will receive different premiums based on learned zone-level risk.

> Note: Premium amounts shown below are illustrative estimates for design purposes only. Final amounts will be determined during development.

| Worker Zone | Risk Level | Weekly Premium (est.) | Max Weekly Payout (est.) |
|---|---|---|---|
| Koramangala, Bengaluru (low flood risk) | Low | Rs. 12 | Rs. 600 |
| Rohini, Delhi (heat zone) | Medium | Rs. 17 | Rs. 700 |
| Tambaram, Chennai (flood-prone) | High | Rs. 22 | Rs. 800 |

---

### Step 4 — Premium Payment Options

Workers will have flexibility in how they pay their weekly premium. The plan includes three modes:

**Automatic deduction:** Premium is deducted from weekly earnings within the platform at the end of each week. No separate action needed.

**Manual payment (default):** Workers can turn on automatic deduction or instead choose to pay their premium at any point during the week — as a single payment or split across multiple days — as long as the total is settled before the week ends. This gives workers who prefer to control their own payment timing the option to do so.

**Advance partial payment:** A worker can pay a lump sum of approximately half the standard weekly premium amount upfront (for example, if the weekly premium is Rs. 20, they may pay Rs. 200 in advance). This advance covers their premium for the corresponding number of weeks that amount covers, plus two additional weeks beyond that at no charge. After this period ends, normal weekly payments resume.

**Non-payment consequences:** If a worker misses a weekly payment, coverage pauses from the following week until the outstanding amount is cleared. Consecutive missed payments for more than two weeks will result in the policy being suspended. A suspended policy requires re-enrollment and a fresh waiting period before claims are eligible again. Workers will be notified before suspension occurs.

---

### Step 5 — Parametric Trigger Monitoring

The system will continuously monitor external data sources and internal platform metrics. When a trigger threshold is crossed, the system logs a disruption event for the relevant zone and notifies affected workers.

> Note: Trigger thresholds listed below are indicative. Final values will be calibrated during development.

| Disruption Type | Indicative Trigger Condition | Data Source |
|---|---|---|
| Heavy Rain | Rainfall above threshold in worker's zone | OpenWeatherMap |
| Extreme Heat | Temperature above threshold during active hours | OpenWeatherMap |
| Severe Pollution | AQI above hazardous level in worker's city | OpenAQ / WAQI |
| Curfew / Bandh | Verified zone closure | Traffic API / Government alert |
| Platform Order Drop | Significant order volume drop sustained over time in zone | InDel internal data |
| Flash Flood | Flood alert issued for zone by IMD | Weather alert API |

Because InDel will own its platform data, the order volume drop trigger will be particularly useful — no external API is needed to detect when deliveries in a zone have effectively stopped.

---

### Step 6 — Income Loss Calculation

When a disruption event is logged, the system will estimate the worker's income loss for that window.

```
Baseline hourly rate  = Average hourly earnings over past 4 weeks (from InDel data)
Disruption window     = Time from disruption start to end
Expected earnings     = Baseline hourly rate x disruption hours
Actual earnings       = Earnings recorded in InDel during disruption window
Income loss           = Expected earnings - Actual earnings
Payout amount         = Income loss x coverage ratio (capped at weekly maximum)
```

> Note: The coverage ratio is an estimated figure for modelling purposes and will be confirmed during the financial design phase.

**Illustrative example:**

A worker earns an average of Rs. 120/hour over 4 weeks. A flood event is logged from 11:40 AM to 5:30 PM (approximately 5 hours 50 minutes).

```
Expected earnings:          Rs. 120 x 5.83 hrs  = Rs. 700
Actual earnings:            2 partial deliveries = Rs. 80
Estimated income loss:      Rs. 700 - Rs. 80     = Rs. 620
Estimated payout (approx.): Rs. 558
```

---

### Step 7 — Claim Filing

Unlike fully automated payout systems, InDel is planned to let the worker decide when to file a claim. When a disruption event has been logged for their zone and they believe they have experienced income loss, the worker files a claim request through their dashboard.

Once filed, the system runs eligibility verification:

- Was the worker logged into InDel during the disruption window?
- Does GPS confirm presence in the affected zone?
- Does InDel order activity confirm reduced earnings during the window?
- Is the claim behavior consistent with the zone-wide claim cluster?

If eligibility is confirmed, the claim is approved and payout is initiated. This approach keeps the worker in control of when they claim while still using automated verification to confirm legitimacy.

---

### Step 8 — Payout

Approved payouts will be sent to the worker via:

- UPI direct transfer
- InDel in-app wallet (can be applied toward future premium payments)
- Bank transfer (next-day settlement for amounts above a defined threshold)

Workers will receive a notification explaining the disruption event, the income loss estimate, and the amount approved. This is a deliberate design choice — transparency about how a payout was calculated builds trust over time.

---

### Step 9 — Maintenance Check (Self-Service Claim Audit)

Workers will have access to a Maintenance Check feature for situations where they believe they should have been eligible for a claim but were not notified or faced a rejected eligibility check.

**Phase 1 — Automated AI Response:**

When a worker triggers a Maintenance Check:

1. The system sends a query to an AI API with the worker's stored activity data, GPS records, disruption signals in their zone, and the SHAP-based explanation of why their claim was or was not flagged as eligible.
2. The response is returned to the worker in plain language, explaining what signals were detected, what the model assessed, and what — if anything — may have gone wrong.
3. The worker can review this output and, if they believe there is an error, escalate to insurer review with the explanation as supporting context.

**Phase 2 — Human Reviewer Follow-Up:**

After the AI response is delivered, the Maintenance Check is also logged in the insurer's admin queue. A designated maintenance reviewer on the insurer's side will independently examine the same data — the worker's activity records, the disruption signals, the AI output, and the SHAP breakdown — to check whether there is a systemic issue with the model's assessment or a data gap that caused an incorrect result.

The reviewer will send a follow-up message to the worker confirming either that the AI explanation was accurate, that a correction has been made, or that the issue has been identified and escalated for a policy-level fix. This message is also delivered in the worker's preferred language.

This two-step process means the worker is not left relying solely on an automated response. Every Maintenance Check has a human confirmation loop attached to it, which also serves as a feedback mechanism for the insurer to catch and fix model errors over time.

**Design constraints for Maintenance Check:**
- Limited to 3 uses per day per worker to prevent unnecessary API load
- Available in all major Indian languages — both the AI response and the human follow-up message will be delivered in the worker's preferred language as set during onboarding
- The check does not automatically approve a claim — it provides an auditable explanation that both the worker and insurer can reference
- Human reviewer follow-up is targeted within a reasonable response window, to be defined during development

This feature gives workers visibility into system decisions in a language they are comfortable with, and ensures there is always a human accountability layer behind the automated output.

---

## Weekly Premium Model

**Planned base structure:**

> Note: All figures below are illustrative estimates. The hackathon guidelines state that specific amounts do not need to be finalised at this stage and will be determined during development.

- Weekly premium range: Rs. 10 — Rs. 25 (dynamically calculated per worker per zone)
- Coverage ratio: Estimated at 80–90% of calculated income loss
- Maximum weekly payout: Estimated at Rs. 800
- Premium deduction: Automatic by default, with manual and advance payment options available

**Continuity rewards:**

Workers who maintain consistent premium payments without claims over time will receive incremental benefits — reduced premiums in upcoming weeks, extended coverage periods, or increases to their maximum payout ceiling. The specific milestones and reward values will be defined during the product design phase. The intent is to make consistent enrollment worthwhile without creating a system that feels like it is withholding something until certain conditions are met.

**Non-payment consequences:**

Missed payments pause coverage from the following week. Two or more consecutive missed weeks result in policy suspension, requiring re-enrollment and a fresh waiting period before claims are valid again.

---

## Planned AI and ML Integration

A core design principle of InDel is that parametric triggers are threshold conditions that initiate a disruption log, but the AI layer sits above them to determine risk pricing, verify claim legitimacy, and forecast future exposure. The system is not intended to be a rule engine — thresholds are inputs to ML models, not the decision-makers themselves.

---

### Model 1 — Dynamic Premium Calculation (XGBoost Regressor with SHAP Explainability)

Will predict the expected weekly income loss probability for a given worker profile and zone.

**Planned input features:**
- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours on InDel
- Platform order density variance in the zone (InDel internal)
- Worker income stability score (earnings variance over past 8 weeks)

**Training data:** Synthetic dataset from IMD historical weather records, CPCB AQI archives, and simulated InDel order disruption logs during the early phase. A continuous background retraining pipeline will update the model as real platform data accumulates — no manual intervention required.

**Output:** Continuous risk score (0–1) mapped to weekly premium.

**Retraining cadence:** Monthly in prototype; continuous pipeline in production.

**Explainability — SHAP Integration:**
The model will use SHAP (SHapley Additive exPlanations) values to make premium calculations auditable. Each worker's premium will be explainable in terms of contributing factors:

```
Premium = Rs. 18 because:
  Flood risk in zone        +Rs. 6
  Rolling AQI pattern       +Rs. 3
  Income instability score  +Rs. 2
  Base rate                  Rs. 7
```

This breakdown is used in the Maintenance Check feature and is available to insurers for regulatory and audit purposes.

---

### Model 2 — Fraud Detection (Isolation Forest + DBSCAN Clustering + Rule Overlay)

A three-layer stack where each layer catches a different class of fraudulent behavior.

**Layer 1 — Isolation Forest anomaly detection:**

Trained on expected claim behavior patterns across the worker pool. Flags workers whose claim profile deviates statistically from the zone-wide cluster.

Planned input features:
- GPS trail consistency during disruption window
- Ratio of claimed loss to historical earnings baseline
- Claim frequency per worker over rolling 8-week window
- Zone-wide claim clustering (are other workers in the same zone also claiming?)
- Mobility pattern score (how stable is the worker's operating zone week over week)

**Layer 2 — DBSCAN Cluster-Based Behavior Analysis:**

Workers in the same zone during the same disruption should show similar claim patterns. DBSCAN groups workers into behavioral clusters per disruption event. Workers who fall outside any cluster — noise points — are flagged for review. This catches fraud that Isolation Forest alone can miss: a worker whose individual claim history looks normal but whose behavior during a specific event diverges from everyone else in the same zone at the same time.

**Layer 3 — Rule overlay for hard disqualifiers:**
- Worker GPS not in the affected zone at trigger time: auto-reject
- InDel platform shows completed deliveries during the claimed disruption window: auto-reject
- Anomaly score above threshold from Layer 1: route to manual review queue

Each layer is independent. A claim must clear all three to be approved.

---

### Model 3 — Disruption Forecasting (Facebook Prophet — Time Series)

A forward-looking model forecasting likely disruption events and associated claim volume for the coming week, broken down by zone. Feeds the insurer dashboard for reserve planning.

**Planned input:** Historical weather, AQI trends, InDel order volume history by zone and season.

**Output:** Zone-level claim probability for next 7 days.

**Use:** Insurer reserve planning only — not used for individual claim decisions.

**Retraining cadence:** Weekly.

Prophet is selected for the prototype because it is reliable on small datasets and handles seasonal patterns well. As real data accumulates, this model is planned for upgrade.

---

### Model Card Summary

| Model | Type | Primary Input | Output | Retraining |
|---|---|---|---|---|
| Premium Calculator | XGBoost + SHAP | Zone risk features + worker profile | Weekly premium + breakdown | Monthly / continuous |
| Fraud Detector | Isolation Forest + DBSCAN + Rules | GPS + claim behavior + InDel activity | Anomaly score + cluster verdict + decision | Weekly |
| Disruption Forecaster | Prophet Time Series | Historical weather + InDel order logs | Zone claim probability | Weekly |

---

### Future Model Upgrades

As InDel accumulates real historical data, the forecasting component is planned for upgrade:

**DeepAR** — A probabilistic deep learning model that learns correlated disruption patterns across multiple zones simultaneously. Unlike Prophet which handles each zone independently, DeepAR will learn that when zone A shows a rainfall-driven order drop, nearby zone B is likely to follow. Planned once at least 6 months of real zone-level data is available.

**Temporal Fusion Transformer (TFT)** — A state-of-the-art multi-horizon forecasting model that combines attention mechanisms with recurrent networks and can incorporate multiple time-varying features simultaneously. Outputs interpretable attention weights showing which inputs drove each forecast. Planned as a longer-term upgrade once compute infrastructure is in place.

| Component | Prototype | Production |
|---|---|---|
| Risk Pricing | XGBoost + SHAP | Same, larger training corpus |
| Fraud Detection | Isolation Forest + DBSCAN + Rules | Same stack, larger corpus |
| Disruption Forecasting | Prophet | DeepAR → TFT |

---

## Illustrative Unit Economics

Conservative assumptions for a cohort of 1,000 active workers in Chennai during a standard month. These figures are used to validate that the financial model is directionally viable, not as final projections.

| Assumption | Value |
|---|---|
| Average weekly premium | Rs. 17 |
| Active weeks per month | 4 |
| Disruption events per worker per month | 0.8 |
| Average payout per event | Rs. 550 |

| Metric | Value |
|---|---|
| Total premium collected | Rs. 68,000 |
| Expected total payouts | Rs. 44,000 |
| Gross margin before ops cost | Rs. 24,000 (35%) |
| Projected loss ratio | ~65% |

A 65% loss ratio is within the acceptable range for microinsurance. Standard health microinsurance in India operates at 70–85%. InDel's parametric structure should keep loss ratios predictable because payouts are capped and calculated algorithmically.

| City | Risk Profile | Avg. Weekly Premium | Expected Monthly Loss Ratio |
|---|---|---|---|
| Chennai | High (monsoon + heat) | Rs. 22 | 72% |
| Bengaluru | Medium | Rs. 16 | 61% |
| Pune | Low | Rs. 11 | 54% |

---

## Scenario Walkthroughs

These scenarios illustrate how the system is designed to behave. They are based on current design assumptions and will be validated through simulation during development.

**Scenario 1 — Flood Event (Chennai, August)**
Worker: InDel rider, Tambaram, earns Rs. 4,200/week, premium Rs. 22.
Event: Heavy rainfall logged above threshold at 11:40 AM.
Flow: System logs disruption in Tambaram zone. Worker files claim via dashboard at 1:00 PM. Eligibility check: GPS confirms zone presence, InDel confirms order drop, fraud check passes. Payout of Rs. 360 (estimated 6-hour window loss) approved. Worker receives payment via UPI shortly after.

**Scenario 2 — Heat Wave (Delhi, May)**
Worker: InDel rider, Rohini, earns Rs. 3,800/week, premium Rs. 19.
Event: Temperature threshold crossed at 1:00 PM.
Flow: Disruption logged. Worker files claim. InDel activity shows significant order drop. GPS in zone confirmed. Payout for 4-hour window estimated at Rs. 270, approved.

**Scenario 3 — Continuity Reward (Pune, February)**
Worker: InDel rider, Kothrud, earns Rs. 3,500/week, premium Rs. 11.
Event: Several consecutive weeks without a claim.
Flow: System applies continuity reward — reduced premium or extended coverage period depending on milestone reached. Worker notified.

**Scenario 4 — Transit Disruption (Mid-Delivery Flood)**
Worker: InDel rider, active delivery from Adyar to Velachery, Chennai.
Event: Flash flood logged in Guindy (between Adyar and Velachery) at 3:15 PM.
Flow: InDel confirms active order at 3:15 PM. GPS shows rider stopped in Guindy at 3:18 PM. Flood logged in Guindy at 3:12 PM. All four transit verification conditions satisfied. Worker offline — claim is queued. When worker reconnects, they are notified of the disruption event and can file the claim. Payout processed on filing.

**Scenario 5 — Zone Hopping Attempt (Fraud Caught)**
Worker: InDel rider, Pune (low risk), premium Rs. 11.
Event: Worker moves GPS to Chennai flood zone the day before a major rainfall event.
Flow: Mobility model detects unusual zone shift with no Chennai activity history. Anomaly score high. Zone lock active — Chennai claims ineligible for 7 days. If claim is filed, it is auto-rejected. Premium auto-adjusts to Chennai risk profile from next cycle.

**Scenario 6 — Maintenance Check (Worker Disputes Eligibility)**
Worker: InDel rider, believes they should have been eligible for a claim but system did not flag it.
Flow: Worker triggers Maintenance Check from dashboard. System calls AI API with worker's activity data, zone disruption signals, and SHAP breakdown of the eligibility model's assessment. Response is returned in Tamil (worker's preferred language) explaining what was detected and why the eligibility check did not trigger. Simultaneously, the check is logged in the insurer's admin queue. A maintenance reviewer examines the same data independently and sends a follow-up message to the worker in Tamil — confirming the AI explanation was accurate, or noting that a correction has been made, or flagging a model issue for escalation. Worker receives both messages and decides whether further action is needed.

**Scenario 7 — National Lockdown**
Event: National lockdown. 78% of InDel workers across all zones file claims in one week.
Flow: Aggregate claims exceed pool threshold. Catastrophic Cap activated. Individual payouts reduced proportionally. Workers notified. Reinsurance layer activated for insurer. From week 3: Lockdown Partial Coverage Clause applies — reduced payout rate, premiums suspended.

---

## Risk Controls and Edge Cases

This section documents how InDel is designed to handle failure modes and adversarial scenarios.

---

### Edge Case 1 — Global Lockdown or Mass Correlated Disruption

**The problem:** A large-scale event hits every worker simultaneously. The premium pool risks depletion in a single week — correlated risk is the primary actuarial failure mode for parametric insurance at scale.

**Our design approach:**

A Catastrophic Event Cap will activate when aggregate claims exceed a defined percentage of the active premium pool in a single week. Individual payouts will be proportionally reduced — workers receive less, not nothing.

Formula: Individual payout = Calculated entitlement x (Available pool / Total eligible claims)

A Reinsurance Layer is modelled into the insurer deployment architecture. The deploying insurer purchases reinsurance activating when weekly aggregate claims exceed a set threshold. This is included in the financial model to demonstrate production viability, not in the hackathon prototype.

A Lockdown Partial Coverage Clause will define government-mandated full lockdowns as a special disruption category. Coverage will be capped at a reduced rate for up to 2 consecutive weeks. Beyond that, coverage pauses and premiums are suspended. This is disclosed at onboarding.

---

### Edge Case 2 — Zone Hopping (Deliberate Location Fraud)

**The problem:** A worker enrolls in a low-risk zone, then relocates to a high-risk zone before a disruption to claim a payout they underpaid for.

**Our design approach:**

Zone Lock with Cooling Period: When GPS detects a zone change, the new zone's risk profile immediately applies to premium calculation. A 7-day waiting period is enforced before claims in the new zone are eligible.

Mobility Pattern Scoring: The fraud model will include zone-change frequency as a feature. Workers who suddenly appear in a high-risk zone with no prior activity history there will receive a high anomaly score.

Premium Auto-Adjustment: If GPS activity consistently shows the worker outside their declared zone over a rolling 2-week period, the system will reclassify their risk profile to reflect actual operating location.

---

### Edge Case 3 — Transit Disruption (Disruption Between Delivery Points)

**The problem:** Worker is mid-delivery when a disruption occurs between their start and destination. They may have no connectivity. Their enrolled zone differs from the disruption location.

**Our design approach:**

Transit Disruption Events will be a distinct claim type. The enrolled zone is irrelevant — the coverage anchor is the active InDel delivery order.

Verification uses four conditions:
- Active InDel delivery order existed at the time of disruption
- GPS trail shows directional movement consistent with the delivery route before stoppage
- The disruption zone had a verified trigger active at the time of GPS stoppage
- GPS stoppage occurred after the trigger fired, not before

If all four conditions are met, the system flags the claim as eligible. When the worker reconnects, they are notified and can file. Zone-lock and home-zone rules do not apply to Transit Disruption Events.

Scalability consideration: During mass events, the system falls back to zone-cluster verification rather than individual route tracing. Individual route analysis is reserved for anomaly-flagged claims only.

---

### Edge Case 4 — Interstate Travel

**The problem:** A worker's insurance is calibrated for their home state. Travel elsewhere puts the risk model outside its training data.

**Our design approach:**

Home Zone Anchor with Portable Coverage: Coverage travels with the worker for up to 72 hours in another state using home zone parameters.

Zone Migration for Extended Stays: Beyond 72 hours, the system flags a zone migration event. The worker is prompted to update their registered zone. A 7-day waiting period applies before claims in the new zone are valid. Premium is recalculated at the next weekly cycle.

Interstate Transit Disruptions follow the Transit Disruption Event logic — state boundaries do not affect coverage when the worker is mid-delivery on an active InDel order.

---

### Edge Case Summary

| Scenario | Detection Method | Planned System Response |
|---|---|---|
| Global lockdown / mass event | Aggregate claims exceed pool threshold | Proportional payout reduction + reinsurance activation |
| Zone hopping | Mobility anomaly score + GPS zone mismatch | 7-day zone lock + premium auto-adjustment |
| Mid-delivery transit disruption | Active order + GPS trail + trigger timing | Claim flagged as eligible, worker files on reconnection |
| Interstate travel under 72 hours | GPS state detection | Home zone rules apply, coverage continues |
| Interstate travel over 72 hours | Persistent GPS state mismatch | Zone migration prompt + 7-day waiting period |
| Connectivity loss during disruption | Worker offline | Disruption logged, worker files claim on reconnection |

---

## Planned Dashboards

### Worker Dashboard
- Active coverage status and current weekly premium
- Earnings this week vs estimated protected income
- Disruption alerts active in their zone
- Claim history and wallet balance
- Continuity reward progress
- Maintenance Check access (up to 3 uses per day) — shows AI response and human reviewer follow-up message when available
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
- Maintenance Check review queue — pending worker checks requiring human reviewer follow-up, with full AI output and worker data visible for each

---

## Compliance and Regulatory Considerations

**Product Classification:** Parametric income protection falls under general insurance. The deploying insurer would file with IRDAI as a group microinsurance policy, which carries a simplified approval pathway.

**Data Privacy:** Worker data will fall under the Digital Personal Data Protection Act 2023. The planned architecture separates PII from risk modelling inputs and will not store raw GPS trails beyond the claim verification window (72 hours post-disruption).

**Consent:** Insurance enrollment is opt-in. Workers can pause or cancel coverage at any time. Premium deductions require explicit consent at enrollment.

**Payout Classification:** Payouts are compensation for income loss, not indemnity for an insured asset. Payouts below Rs. 2,50,000 annually are unlikely to create tax obligations for gig workers at current income levels.

**Language Support:** The platform and all worker-facing communications — including Maintenance Check outputs — are planned to support all major Indian languages via a translation layer. This is a core accessibility requirement, not an optional feature.

Note: A production deployment would require the deploying insurer to handle IRDAI product registration and KYC/AML obligations. These are outside the scope of the hackathon prototype.

---

## Tech Stack

| Layer | Planned Technology |
|---|---|
| Backend | Python (FastAPI) |
| Frontend | React.js |
| Database | PostgreSQL |
| AI / ML (Prototype) | scikit-learn, XGBoost, SHAP, Prophet, DBSCAN |
| AI / ML (Future) | DeepAR, Temporal Fusion Transformer |
| Weather API | OpenWeatherMap (free tier) |
| AQI API | OpenAQ / WAQI |
| Traffic / Zone Alerts | Mock API (simulated) |
| Payment | Razorpay test mode / UPI simulator |
| Hosting | AWS / Render |

---

## Why This Approach

Parametric insurance systems require reliable activity data to verify income loss events. In many real-world implementations, this data is owned by external delivery platforms, creating a dependency that makes verification difficult and fraud detection unreliable.

InDel addresses this limitation by integrating delivery operations and the insurance engine into a single platform architecture. Because worker activity, order patterns, and earnings history are recorded directly within the system, the insurance layer always has access to verifiable first-party data.

This enables more accurate income loss estimation, stronger fraud detection, and faster claim verification compared to systems that depend on external data pipelines.

The worker-initiated claim model also ensures that payouts occur only when the worker confirms they experienced income loss, while automated verification ensures that eligibility checks remain consistent and scalable.

Finally, the Maintenance Check feature introduces transparency into the system. Workers can request an explanation of how a claim decision was reached in their preferred language, allowing them to understand and audit the system’s reasoning without relying solely on customer support channels.

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
