# InDel — Frequently Asked Questions

> This document anticipates questions judges, insurers, and stakeholders may raise about InDel's design, business model, research, and feasibility. Answers reflect the thinking behind each decision, not just the decision itself.

---

## On the Data Model

**Q: How do you get first-party delivery data without building a full consumer marketplace?**

InDel operates as a white-label delivery management layer that integrates with an existing platform via API. The platform partner (e.g. Swiggy, Zomato) routes orders to InDel-enrolled workers. InDel handles assignment, GPS tracking, and earnings recording on top of that order flow. We don't generate consumer demand — we sit inside the operational layer that already exists. The delivery data is ours because we handle the worker-side of the transaction, not because we built a consumer app.

---

**Q: What happens if Swiggy/Zomato won't integrate with you?**

This is exactly the problem every previous parametric insurance attempt ran into — and it's why InDel is structured differently. We don't ask platforms to share data. We offer them something: their delivery workforce becomes financially protected without them having to build any insurance infrastructure themselves. That's a benefit to the platform, not a request for a favour. In zones where no platform partner exists, InDel can operate as a standalone delivery management system for insurer-deployed worker cohorts — a secondary use case, but a real fallback.

---

**Q: How is your data ownership different from what competitors have tried?**

Most parametric insurance attempts for gig workers assume they can get activity data from Swiggy or Zomato after the fact — via API agreements, data sharing deals, or scraping. Those platforms have no incentive to share, and their TOS actively prohibits third-party API access. One of our own research respondents raised this unprompted: "Any optimization would require API/handshake from various apps which is most likely against TOS." InDel doesn't request data from platforms. We generate it ourselves by being part of the delivery operation. That's the structural difference.

---

## On Fraud

**Q: Why is your fraud detection simpler than other solutions?**

Because we have a smaller fraud problem to begin with. Other systems need complex multi-layer fraud detection — device attestation, GPS triangulation, Cell-ID cross-checks — because they don't own the activity data and have to infer whether a worker was genuinely impacted. InDel has the earnings record directly. A worker either completed orders on InDel during the disruption window or they didn't. That's a first-party fact, not an inference. Complexity in fraud detection is the price of not owning your data source. Our architecture eliminates the attack surface rather than defending against it.

---

**Q: What stops a worker from just not accepting orders during a disruption to fake income loss?**

Several things simultaneously. First, the disruption must be independently confirmed by external signals — weather, AQI, zone closure — before any claim window opens. A worker cannot create a disruption by refusing orders. Second, the income loss calculation is based on zone-wide earnings patterns, not individual behaviour. If a worker's earnings drop but their zone peers' earnings didn't, that divergence is flagged. Third, the Isolation Forest model detects individual claim profiles that deviate statistically from zone-wide patterns during the same event. A worker faking income loss looks different from a zone of workers who genuinely experienced it.

---

**Q: How do you handle GPS spoofing?**

InDel's fraud model is built around economic activity validation, not location validation. The question we ask is not "was the worker in the zone?" but "did the worker experience an income loss consistent with other workers in the same zone?" A spoofed GPS position doesn't help a fraudster if their earnings pattern doesn't match the disruption. Additionally, the rule overlay layer auto-rejects claims where InDel's own platform shows completed deliveries during the claimed disruption window — you cannot simultaneously be "unable to work" and completing orders on the same system.

---

## On the Business Model

**Q: Who is your actual customer — the insurer or the worker?**

The insurer. InDel is a B2B platform. Workers are the end beneficiaries and the product surface, but the insurer is the paying customer. This is important because it means InDel never handles premium pools, never underwrites risk, and never takes on actuarial liability. We provide the infrastructure. The insurer deploys it, owns the policy, and manages the capital. This is the same model Guidewire uses — infrastructure for insurers, not insurance for consumers.

---

**Q: How does InDel make money without taking a cut of premiums?**

SaaS-style platform fee per active worker per month, charged to the insurer. Taking a cut of premiums would create a misaligned incentive — InDel would benefit from minimising claim approvals. A fixed platform fee means InDel's revenue is independent of claim outcomes, which keeps our incentives aligned with accurate processing. A small per-approved-claim processing fee recovers the marginal cost of running fraud detection and payout infrastructure.

---

**Q: What's your loss ratio assumption and how did you arrive at it?**

We modelled a projected loss ratio of approximately 65% for a Chennai cohort — within the acceptable range for microinsurance. Standard health microinsurance in India operates at 70–85%. Our figure is conservative because parametric income insurance has a narrower trigger set than health insurance — not every bad day qualifies, only verified disruption events with confirmed income impact. The 65% figure was stress-tested against three cities with different risk profiles: Chennai at 72% (high monsoon and heat exposure), Bengaluru at 61% (medium risk), and Pune at 54% (low risk). These are illustrative estimates, not actuarial projections — final figures would be refined with an insurer partner.

---

## On Adoption

**Q: Why would workers pay for this when they currently don't have insurance?**

Two reasons. First, the premium is positioned as the cost of one to two deliveries per week — not an abstract financial product but a concrete trade-off workers already understand. Second, and more importantly, the premium is collected at the weekly earnings payout moment — the highest trust, lowest resistance point in the worker's financial cycle. The worker has just received money. A small prompt at that moment, similar to how delivery apps present tip options at order completion, frames the premium as an allocation from money just received rather than money leaving the account. Adoption in microinsurance fails when payment feels like a loss. InDel's timing removes that feeling.

---

**Q: How do you collect premiums from workers with irregular income?**

Three payment options address this directly. Automatic deduction from weekly earnings at payout — the default and lowest friction option. Manual payment at any point during the week. And advance partial payment — a lump sum covering multiple weeks, which suits workers who have a good week and want to lock in coverage. If a worker misses a week, coverage pauses rather than the policy cancelling immediately. Two consecutive missed weeks triggers suspension with a fresh waiting period on re-enrollment. This mirrors how gig workers actually manage money — in variable weekly cycles, not fixed monthly commitments.

---

**Q: Why would insurers trust a student-built platform with their policy infrastructure?**

They wouldn't — not immediately, and we're not claiming they would. InDel at hackathon stage is a proof of concept for the infrastructure model, not a production deployment. The value we're demonstrating is that this architecture is possible, that the unit economics work, and that the data ownership problem — which has killed every previous attempt — has a viable solution. A real deployment would involve the insurer handling IRDAI product registration and KYC/AML obligations, with InDel as the technology layer underneath. We validated this framing through outreach to insurance sector contacts during our research phase.

---

## On Research

**Q: Did you talk to real delivery workers?**

Yes — through multiple channels. We spoke directly with active delivery partners during their routes in our local area, keeping conversations to 3–5 minutes out of respect for their time. No recordings were made — workers in the middle of a delivery shift are not in a position to consent to being interviewed on camera, and asking would have ended the conversation. We also conducted structured research via Reddit (r/AmazonFlexDrivers), reaching 892 delivery workers with direct questions about income disruption and willingness to pay.

---

**Q: Did you validate willingness to pay?**

Yes — and we got honest pushback, which shaped the design. One respondent told us directly: "Bad weather is when I make the most money. I can count on one hand how many days a year deliveries are actually stopped. I'd rather put ₹20–₹30 a week in savings." That response influenced three specific design decisions: keeping premiums low because genuine full stoppages are infrequent, building multi-signal validation to distinguish weather that slows deliveries from weather that stops them entirely, and calibrating payouts to verified income loss rather than weather events alone. A product that only collected confirmatory responses would be weaker for it.

---

**Q: Did you talk to any insurers?**

We attempted to get an appointment with insurance sector contacts to validate the B2B model and the financial assumptions. This informed the loss ratio modelling and the SaaS fee structure. A production deployment would require deeper insurer partnership — we are not claiming otherwise.

---

## On Build Feasibility

**Q: Is this too complex to actually build?**

The full spec is intentionally comprehensive for an ideation document. The prototype is scoped to a single demonstrable loop: worker onboarding → disruption trigger → income loss calculation → fraud check → payout → insurer dashboard update. Every component in that loop uses established, well-documented technology — FastAPI, PostgreSQL, scikit-learn, Kafka, Razorpay sandbox. The more complex features — DeepAR forecasting, full SHAP explainability, the Maintenance Check AI layer — are explicitly marked as future scope. The tech stack note in the README states clearly that planned technologies may change during development.

---

**Q: What are you building in 4 weeks versus what's future scope?**

**Building:**
- Worker onboarding and zone assignment
- Event-driven disruption trigger (weather API integration)
- Income loss calculation engine
- Isolation Forest fraud detection
- Automated payout via Razorpay sandbox
- Insurer dashboard — loss ratio, fraud queue, premium pool health, 7-day forecast

**Future scope:**
- DBSCAN cluster fraud layer
- Full SHAP explainability in worker-facing language
- Maintenance Check AI feature
- DeepAR forecasting upgrade
- Transit disruption edge case handling
- Interstate travel rules

The prototype demonstrates the core value proposition end to end. Everything else is an enhancement on top of a working foundation.

---

*Document maintained by Team ImaginAI — Guidewire DEVTrails 2026*

screenshots:
<img width="1512" height="803" alt="image" src="https://github.com/user-attachments/assets/5c5a126a-77e8-4472-a370-9540b7ab6994" />
<img width="1512" height="803" alt="image" src="https://github.com/user-attachments/assets/61915ae3-c490-48f9-9a23-f6d8afbac644" />

