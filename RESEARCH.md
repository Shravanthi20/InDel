## User Research & Problem Validation

**Method:** Primary research conducted via Reddit (r/AmazonFlexDrivers) — structured interview questions posted by the InDel team, responses from active delivery partners.

**Post reach:** 892 views across the thread — high reach relative to a niche delivery worker community.

**Why low comment rate matters:** Gig workers are reluctant to publicly discuss income vulnerability in communities where platform employers may be present. The low comment-to-view ratio is itself a product insight — these workers need a protection system that works silently and automatically, without requiring them to surface financial distress publicly. InDel's zero-touch automatic payout design is a direct response to this reality.

---

**Key finding 1 — The data ownership problem (unprompted):**
One respondent independently raised the API/TOS barrier without being prompted:

> *"Any optimization would require API/handshake from various apps which is most likely against TOS"*

This is an independent third-party validation of InDel's core architectural decision — the white-label integration model exists precisely because workers themselves recognise that third-party data access is structurally blocked.

---

**Key finding 2 — Survival mode framing (unprompted):**

> *"It's more a survival mode than an optimization mode"*

This single phrase captures the emotional reality of InDel's target user more precisely than any statistic. Workers aren't optimizing income — they're protecting it from collapse. InDel's design philosophy reflects this directly.

---

**Key finding 3 — Honest pushback that shaped the design:**
When asked directly: *"If there was a service that automatically paid ₹300–₹500 when events like heavy rain, floods, curfews, or app outages stopped deliveries, would you be willing to pay a small weekly fee (₹20–₹30) for it?"*

One respondent pushed back:

> *"Bad weather is when I make the most money. I can count on one hand how many days a year deliveries are actually stopped. I'd rather put ₹20–₹30 a week in savings."*

**This response directly influenced InDel's design in three ways:**
- Dynamic pricing model — premiums are low precisely because genuine full stoppages are infrequent
- Multi-signal validation — the system distinguishes between weather that slows deliveries and weather that stops them entirely
- Coverage calibration — payouts trigger only on verified income loss, not on weather events alone

---

screenshots:
<img width="1512" height="803" alt="image" src="https://github.com/user-attachments/assets/5c5a126a-77e8-4472-a370-9540b7ab6994" />
<img width="1512" height="803" alt="image" src="https://github.com/user-attachments/assets/61915ae3-c490-48f9-9a23-f6d8afbac644" />

