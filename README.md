# InDel: Insure,Deliver — Parametric Income Protection for Delivery Partners

**Team:** ImaginAI
**Hackathon:** Guidewire DEVTrails 2026
**Persona:** Food Delivery Partners

---

## The Problem

India's platform-based delivery partners earn based on active working time and completed orders. When external disruptions occur — heavy rain, extreme heat, severe pollution, curfews, or platform outages — deliveries stop and income drops immediately. Workers lose 20–30% of monthly earnings during these events with no financial safety net.

Traditional insurance does not address this. It covers accidents, vehicles, and health. It does not cover the most real risk a delivery worker faces: losing a day's wages because the world made it impossible to work.

---

## What We Are Building

InDel is a **parametric income protection engine** designed to be deployed by insurance companies for platform-based gig workers.

It is not a standalone consumer app. It is B2B infrastructure — a white-label system that insurers can integrate into their product portfolio and distribute through delivery platforms like Swiggy, Zomato, and Blinkit.

When a defined external disruption crosses a measurable threshold, the system automatically triggers a payout to the affected worker. No claim forms. No manual approval. No delay.

---

## The Three Stakeholders

**Insurer (Primary Customer)**
Purchases and deploys InDel as a new product vertical. Gains access to a previously uninsured, high-volume customer segment with data-driven risk modelling and automated claims processing.

**Delivery Platform (Distribution Channel)**
Integrates InDel into their existing rider dashboard. Can offer it as an opt-in benefit or bundle it into rider packages for retention. No additional app required for the worker.

**Delivery Partner (End Beneficiary)**
Receives automatic income compensation when disruptions occur. Experiences insurance as a background safety net, not a product they have to actively manage.

---

## How It Works

### 1. Worker Onboarding
The worker opts in through the delivery platform dashboard with a single action. They provide their working zone, typical hours, and delivery platform. This data seeds their risk profile.

### 2. AI-Powered Risk Profiling
A machine learning model calculates a dynamic weekly premium based on:
- Historical weather and flood data for the worker's zone
- Pollution and AQI patterns
- Local disruption history (curfews, protests, zone closures)
- Platform-level order density patterns

Workers in lower-risk zones pay less. Workers in historically flood-prone zones pay more. The premium is recalculated every week.

**Example:**
| Worker Zone | Risk Level | Weekly Premium |
|---|---|---|
| Koramangala, Bengaluru (low flood risk) | Low | Rs. 12 |
| Tambaram, Chennai (flood-prone) | High | Rs. 22 |

### 3. Parametric Trigger Monitoring
The system continuously monitors external data sources. When a trigger threshold is crossed, a claim event is automatically initiated — no worker action required.

**Defined Triggers:**
| Disruption | Trigger Condition |
|---|---|
| Heavy Rain | Rainfall > 35mm in 3 hours in the worker's zone |
| Extreme Heat | Temperature > 43 degrees C during active hours |
| Severe Pollution | AQI > 300 (Hazardous) in the worker's city |
| Curfew / Bandh | Verified zone closure via traffic or government API |
| Platform Outage | Order volume drop > 80% for > 2 hours in zone |

### 4. Fraud Detection
Because the system integrates with delivery platform data, fraud detection is grounded in real activity signals rather than self-reported claims.

Verification checks include:
- Was the worker logged into the platform during the disruption window?
- Did their GPS location match the affected zone?
- Was order volume in that zone verifiably impacted?
- Has the worker submitted an unusually high frequency of claims?
- Are multiple workers from the same zone claiming simultaneously (expected) or is it an isolated outlier pattern (suspicious)?

This cross-referencing removes the need for manual investigation in most cases.

### 5. Automatic Payout
Once the trigger is verified and fraud checks pass, the worker receives compensation via UPI, in-app wallet credit, or direct bank transfer. The target payout window is under 10 minutes from trigger detection.

**Example Payout Scenario:**
A Zomato rider in Chennai typically earns Rs. 4,200 per week. A cyclone warning causes delivery shutdowns for one full day. The weather API detects rainfall above threshold. Platform data confirms order volume dropped 90% in the zone. The system calculates one day's estimated income loss (approximately Rs. 600) and initiates payout automatically.

---

## Weekly Premium Model

The financial model operates on a weekly cycle to match the earning rhythm of gig workers.

**Base Structure:**
- Weekly premium range: Rs. 10 — Rs. 25 (dynamically calculated)
- Maximum weekly payout: Rs. 800
- Payout calculation: Based on verified disruption duration and historical average daily earnings

**Loyalty Mechanics (Retention Layer):**
To address the psychological barrier of paying premiums without seeing claims, the system includes:
- No Claim Bonus: After 8 consecutive weeks without a claim, the worker earns Rs. 50 wallet credit applicable to future premiums
- Premium Holiday: After 12 consecutive weeks, one week's premium is waived while coverage continues
- Loyalty Multiplier: Workers subscribed for 6+ months receive a 10% payout increase

This maintains premium flow for the insurer while giving workers a tangible reason to stay enrolled.

---

## AI and ML Integration

A key design decision in InDel is that the parametric triggers are threshold conditions that *initiate* a claim event, but the AI layer sits above them to determine risk pricing, verify legitimacy, and predict future exposure. The system is not a rule engine. The thresholds are inputs to ML models, not the decision-makers themselves.

**Model 1 — Dynamic Premium Calculation (XGBoost Regressor)**

Predicts the expected weekly income loss probability for a given worker profile and zone. Output feeds directly into premium pricing.

Input features:
- Zone-level historical disruption frequency (past 24 months)
- Seasonal risk score (monsoon proximity, heat wave history)
- Rolling 4-week AQI average for the zone
- Worker's average daily active hours
- Platform-reported order density variance in the zone

Training data: Synthetic dataset generated from IMD (India Meteorological Department) historical weather records, CPCB AQI archives, and simulated platform disruption logs.

The model outputs a continuous risk score between 0 and 1. This score is multiplied against a base premium (Rs. 10) and a coverage multiplier to arrive at the final weekly premium. This means two workers in the same city but different zones will receive meaningfully different premiums based on learned zone-level risk — not just a fixed city-wide rate.

**Model 2 — Fraud Detection (Isolation Forest + Rule Overlay)**

Parametric insurance is vulnerable to a specific fraud pattern: workers who are not actually in the affected zone at the time of disruption but claim to be. The fraud model addresses this in two layers.

Layer 1 — Isolation Forest anomaly detection:
Trained on expected claim behavior patterns. Flags workers whose claim profile deviates statistically from the zone-wide claim cluster. Inputs include GPS trail variance during the disruption window, time-of-claim relative to trigger detection, and historical claim frequency per worker.

Layer 2 — Rule overlay for hard disqualifiers:
- Worker GPS not within the declared zone at trigger time: auto-reject
- Claim submitted more than 2 hours after trigger window closed: flag for review
- Worker's platform activity shows active deliveries during the claimed disruption: auto-reject

The separation of layers matters: the ML model catches soft anomalies that rules would miss (e.g., a worker who is technically in the zone but whose behavior pattern looks fabricated). The rule layer handles clear disqualifiers deterministically.

**Model 3 — Disruption Forecasting (Facebook Prophet — Time Series)**

A forward-looking model that ingests historical weather and disruption data to forecast likely claim events for the coming week, broken down by zone. This feeds the insurer dashboard so underwriters can see predicted claim volume before it materialises.

This is not used for individual claim decisions. It is used for portfolio-level risk management — helping the insurer maintain adequate reserves ahead of high-risk periods like monsoon season.

**Model Card Summary**

| Model | Type | Primary Input | Output | Retraining Cadence |
|---|---|---|---|---|
| Premium Calculator | XGBoost Regressor | Zone risk features + worker profile | Weekly premium (Rs.) | Monthly |
| Fraud Detector | Isolation Forest | GPS + claim behavior signals | Anomaly score + decision | Weekly |
| Disruption Forecaster | Prophet (Time Series) | Historical weather + disruption logs | Zone-level claim probability | Weekly |

---

## Unit Economics (Illustrative Model)

The following model uses conservative assumptions for a cohort of 1,000 active workers in Chennai, a high-disruption city, during a standard month.

**Assumptions:**
- Average weekly premium per worker: Rs. 17 (mid-range of Rs. 10–25 band)
- Active weeks per month: 4
- Expected disruption events per worker per month: 0.8 (roughly one event every 5–6 weeks, based on Chennai historical weather data)
- Average payout per event: Rs. 550

**Monthly figures for 1,000 workers:**

| Metric | Value |
|---|---|
| Total premium collected | Rs. 68,000 |
| Expected total payouts | Rs. 44,000 |
| Gross margin before ops cost | Rs. 24,000 (35%) |
| Projected loss ratio | ~65% |

A 65% loss ratio is within acceptable range for microinsurance products. Standard health microinsurance in India operates at 70–85% loss ratios. InDel's parametric structure keeps loss ratios lower because payouts are capped and non-negotiable — there are no inflated claim settlements.

The loyalty mechanics (no-claim bonus, premium holiday) introduce a small retention cost but reduce churn-driven adverse selection, which is the larger actuarial risk.

**Scenario comparison across three city profiles:**

| City | Risk Profile | Avg. Weekly Premium | Expected Monthly Loss Ratio |
|---|---|---|---|
| Chennai | High (monsoon + heat) | Rs. 22 | 72% |
| Bengaluru | Medium | Rs. 16 | 61% |
| Pune | Low | Rs. 11 | 54% |

The dynamic premium model is what keeps these loss ratios stable across different geographies — high-risk zones pay more, which funds higher expected payouts in those zones.

---

## Compliance and Regulatory Considerations

InDel is designed with awareness of India's insurance regulatory framework under IRDAI (Insurance Regulatory and Development Authority of India).

Key considerations:

**Product Classification:** Parametric income protection products are categorised under general insurance. Any insurer deploying InDel would need to file the product with IRDAI as a group microinsurance policy, which has a simplified approval pathway compared to individual policies.

**Data Privacy:** Worker data collected during onboarding (location, earnings, working hours) falls under the Digital Personal Data Protection Act 2023. InDel's architecture separates PII from risk modelling inputs and does not store raw GPS trails beyond the claim verification window (72 hours).

**Consent:** Opt-in is explicit and revocable. Workers can pause or cancel coverage at any time through the platform dashboard. Premium deductions require active consent confirmation at onboarding.

**Payout as Income:** Parametric payouts to workers are compensation for income loss, not indemnity for an insured asset. This distinction matters for tax treatment — payouts below Rs. 2,50,000 annually are unlikely to create tax obligations for gig workers at current income levels.

Note: In the hackathon context, full regulatory compliance is simulated. A production deployment would require the deploying insurer to handle IRDAI product registration and KYC/AML obligations.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Python (FastAPI) |
| AI / ML | scikit-learn, XGBoost, Prophet |
| Weather API | OpenWeatherMap (free tier) |
| AQI API | OpenAQ or WAQI |
| Traffic / Zone Data | Mock API (simulated) |
| Platform Data | Simulated delivery platform API |
| Payment | Razorpay test mode / UPI simulator |
| Frontend | React.js |
| Database | PostgreSQL |
| Hosting | AWS / Render |

---

## Deliverable Scope by Phase

**Phase 1 (Current — Due March 20)**
- This README
- Core idea documentation
- Initial user research insights
- System architecture diagram

**Phase 2 (March 21 — April 4)**
- Worker registration flow
- Dynamic premium calculation engine
- Parametric trigger logic (3–5 triggers)
- Claims management module

**Phase 3 (April 5 — April 17)**
- Advanced fraud detection layer
- Simulated instant payout integration
- Dual dashboard: Worker view + Insurer admin view
- Final pitch deck and demo video

---

## End-to-End Scenario Simulations

**Scenario 1 — Flood city (Chennai, August)**
Worker profile: Zomato rider, Tambaram zone, earns Rs. 4,200/week, weekly premium Rs. 22.
Event: Northeast monsoon causes 48mm rainfall in 3 hours. Trigger threshold crossed at 11:40 AM.
System response: Weather API flags trigger. Platform data confirms 91% order drop in Tambaram. GPS confirms worker was in zone. Fraud check passes. Estimated income loss for disrupted period (6 hours): Rs. 360. Payout initiated at 11:52 AM. Worker receives Rs. 360 via UPI at 11:54 AM.
Total time from trigger to payout: 14 minutes.

**Scenario 2 — Heat city (Delhi, May)**
Worker profile: Swiggy rider, Rohini zone, earns Rs. 3,800/week, weekly premium Rs. 19.
Event: IMD issues heat wave alert, temperature reaches 45 degrees C at 1:00 PM. Trigger threshold crossed.
System response: Temperature API flags trigger. Platform data shows 74% order drop during peak afternoon hours. Worker GPS in zone confirmed. Payout for 4-hour disruption window: Rs. 270 initiated automatically.

**Scenario 3 — Low risk city (Pune, February)**
Worker profile: Blinkit rider, Kothrud zone, earns Rs. 3,500/week, weekly premium Rs. 11.
Event: No disruptions for 8 consecutive weeks.
System response: No-claim bonus of Rs. 50 wallet credit applied automatically at week 8. Worker notified via platform dashboard. Premium for week 9 reduced to Rs. 0 (premium holiday). Coverage continues uninterrupted.

**Scenario 4 — Transit Disruption (Mid-Delivery Flood)**
Worker profile: Swiggy rider, Chennai, active delivery from Adyar to Velachery. Weekly premium Rs. 20.
Event: Flash flood blocks roads in Guindy (between Adyar and Velachery) at 3:15 PM. Trigger threshold crossed.
System response: Platform API confirms active order at 3:15 PM. GPS trail shows rider moving toward Velachery, stops in Guindy at 3:18 PM. Weather API confirms flood trigger in Guindy zone at 3:12 PM. All four transit verification conditions satisfied. Claim auto-approved without worker action. Rider has no connectivity. Payout of Rs. 180 (estimated 2-hour disruption loss) queued. Rider receives payout and notification at 5:40 PM when connectivity restored.

**Scenario 5 — Zone Hopping Attempt (Fraud Caught)**
Worker profile: Blinkit rider, Pune (low risk), weekly premium Rs. 11. Chennai flood season begins.
Event: Worker relocates GPS to Chennai flood zone one day before major rainfall event.
System response: Mobility pattern model detects zone change from Pune to Chennai — a 1,400km shift with no prior movement history in Chennai. Anomaly score: 0.94 (high). Zone lock cooling period active — Chennai claims ineligible for 7 days. Claim auto-rejected. Worker's premium auto-adjusts to Chennai risk profile (Rs. 22) from next weekly cycle.

**Scenario 6 — Global Disruption Cap (COVID-style Lockdown)**
Event: National lockdown announced. 78% of insured workers across all zones trigger claims in the same week.
System response: Aggregate claims exceed 55% pool threshold. Catastrophic Event Cap activated. Individual payouts reduced to 58% of calculated entitlement. Workers notified via platform dashboard. Reinsurance layer activated for insurer. Lockdown Partial Coverage Clause applied from week 3 onward — payouts at 50%, premiums suspended.

---

## Risk Controls and Edge Cases

This section documents how InDel handles failure modes and adversarial scenarios. Most insurance systems are designed for the happy path. The edge cases below represent real situations that would occur at scale and have been addressed in the system design.

---

### Edge Case 1 — Global Lockdown or Mass Disruption Event

**The problem:** When an event like a pandemic lockdown or a citywide flood affects every worker simultaneously, the entire premium pool is at risk of being wiped out in a single week. This is called correlated risk and it is the primary actuarial failure mode for parametric insurance at scale.

**How InDel handles it:**

A Catastrophic Event Cap is applied when aggregate claims in a single week exceed 55% of the active premium pool. When this threshold is crossed, individual payouts are proportionally reduced to ensure the pool survives. Workers are not refused payment — they receive a reduced payout calculated as their individual entitlement multiplied by the pool survival ratio.

Example: If total eligible payouts in a week are Rs. 1,20,000 but the pool only holds Rs. 80,000, each worker receives 66% of their calculated payout.

A Reinsurance Layer is built into the insurer deployment architecture. The deploying insurer purchases reinsurance that activates when weekly aggregate claims exceed 60% of the collected premium pool. This is not implemented in the hackathon prototype but is explicitly modelled in the financial architecture to demonstrate production viability.

A Lockdown Partial Coverage Clause defines government-mandated full lockdowns as a special disruption category. Coverage during lockdowns is capped at 50% of normal payout for up to 2 consecutive weeks. Beyond 2 weeks, coverage pauses automatically and premiums are suspended. This is disclosed to workers at onboarding. Full coverage during a multi-month lockdown is not economically survivable for any microinsurance pool and this clause reflects honest product design.

---

### Edge Case 2 — Zone Hopping (Deliberate Location Fraud)

**The problem:** A worker enrolls in a low-risk zone at a cheaper premium, then physically relocates to a high-risk zone before a known disruption event to claim a higher payout they did not pay for.

**How InDel handles it:**

Zone Lock with Cooling Period: When a worker's declared working zone changes, the new zone's risk profile immediately applies to premium calculation. However, a 7-day waiting period is enforced before claims in the new zone are eligible. This directly eliminates the financial incentive to chase disruptions.

Mobility Pattern Scoring: The fraud detection model includes a zone-change frequency feature. A worker who has operated consistently within a 3km radius for several months and suddenly appears in a flood-affected zone on the day of a trigger is flagged as a statistical outlier. The Isolation Forest model assigns a high anomaly score and routes the claim for manual review.

Premium Auto-Adjustment: If GPS activity data consistently shows the worker operating outside their declared zone over a rolling 2-week period, the system automatically reclassifies their risk profile to reflect where they actually work. Workers cannot maintain a low-risk premium while operating in a high-risk zone.

---

### Edge Case 3 — Transit Disruption (Disruption Between Delivery Points)

**The problem:** A worker is mid-delivery traveling from point A to point B when a flood occurs at point C, which lies between A and B. The delivery is stalled. The worker cannot file a claim because they have no connectivity. Their enrolled zone may be different from where the disruption occurred.

**How InDel handles it:**

This scenario is handled as a distinct claim type called a Transit Disruption Event, separate from standard zone-based claims. The enrolled zone is irrelevant. The coverage anchor is the active delivery order.

Verification is fully automatic using four data points:
- An active platform order existed at the time of disruption, confirmed by platform API
- The worker's GPS trail shows directional movement consistent with the delivery route before the stoppage
- Zone C had a verified disruption trigger active at the time of the GPS stoppage
- The GPS stoppage occurred after the trigger fired in zone C, not before

If all four conditions are satisfied, the claim is auto-approved by the system without any action from the worker. The payout is queued and delivered once the worker regains connectivity. The worker receives a notification explaining what happened and what was paid.

The zone-lock cooling period and home zone rules do not apply to Transit Disruption Events because the worker did not choose to be in zone C. Presence was incidental to an active earning event.

**Scalability note:** Individual route-level GPS verification is computationally expensive at scale. For high-volume events where thousands of workers are mid-delivery simultaneously, the system falls back to zone-cluster verification: was the worker's GPS anywhere within the disrupted zone during the trigger window, did the worker have an active platform session, and did platform-wide order completion rates drop significantly in that zone. Individual route tracing is reserved only for claims flagged as anomalous by the fraud model, keeping the computational load proportional to actual fraud risk rather than total claim volume.

---

### Edge Case 4 — Interstate Travel

**The problem:** A worker's insurance is priced and calibrated for their home state. If they travel to another state, the risk model is operating outside its training data. Additionally, insurers may not hold product licenses in all states, creating a potential coverage gap the moment the worker crosses a state border.

**How InDel handles it:**

Home Zone Anchor with Portable Coverage: The worker's policy is anchored to their registered state at enrollment. Coverage travels with them for up to 72 hours in another state, using the home zone's risk parameters and payout rules. This mirrors how vehicle insurance operates during interstate travel in India.

Zone Migration for Extended Stays: If GPS data shows the worker has been consistently located in a new state beyond 72 hours, the system flags a zone migration event. The worker is notified and prompted to update their registered zone. A 7-day waiting period applies before claims under the new zone are valid. The premium is recalculated for the new state's risk profile at the start of the next weekly cycle.

Regulatory Handling: Interstate coverage portability is managed at the insurer level through a group microinsurance product structure filed with IRDAI, which allows nationwide coverage under a single product registration. Individual state licensing is the responsibility of the deploying insurer and is outside the scope of the InDel platform layer.

Interstate Transit Disruptions follow the same Transit Disruption Event logic described above. State boundaries do not affect coverage eligibility when the worker is mid-delivery on an active order.

---

### Edge Case Summary Table

| Scenario | Detection Method | System Response |
|---|---|---|
| Global lockdown / mass event | Aggregate claims exceed 55% of pool | Proportional payout reduction + reinsurance activation |
| Zone hopping | Mobility pattern anomaly score + GPS mismatch | 7-day zone lock + premium auto-adjustment |
| Mid-delivery transit disruption | Active order + GPS trail + trigger timing alignment | Auto-approved Transit Disruption Event, no worker action needed |
| Interstate travel under 72 hours | GPS state detection | Home zone rules apply, coverage continues |
| Interstate travel over 72 hours | Persistent GPS state mismatch | Zone migration prompt + 7-day waiting period in new state |
| Connectivity loss during disruption | Worker cannot file — system files automatically | Payout queued, delivered on reconnection |

---



Most teams at this hackathon will build a consumer insurance app. InDel is positioned differently — it is insurer-facing infrastructure with a consumer delivery layer on top. This mirrors how real insurance technology actually works and speaks directly to what a B2B company like Guidewire evaluates.

The parametric model eliminates the single biggest operational cost in microinsurance: claims processing. Traditional claims in microinsurance take 3–7 days and require human review. InDel's automated pipeline targets under 15 minutes from trigger to payout with zero human involvement for standard claims. That is an estimated 80–90% reduction in claims processing overhead for the insurer.

The platform integration eliminates the single biggest adoption barrier: distribution. No separate app, no marketing spend per worker, no trust gap — the insurance lives inside the app the worker already uses every day.

Together, these make a product that is financially viable for insurers, operationally simple for platforms, and genuinely useful for workers.

---

## Team ImaginAI

---

*Built for Guidewire DEVTrails 2026 — University Hackathon*
