# InDel SQL Infrastructure & Data Schema

## Overview
This document describes the complete SQL infrastructure for the InDel platform, covering policies, earnings, claims, payouts, and batch delivery workflows.

## Core Data Tables

### Authentication & Users
- **users** (migration 000002)
  - id, phone, email, role
  - Primary auth table

- **zones** (migration 000001)
  - id, name, level (A/B/C), city, state, risk_rating
  - Zone definitions for delivery routing

- **worker_profiles** (migration 000003)
  - worker_id (refs users.id) ✓ Indexed
  - name, zone_id (refs zones.id), vehicle_type, upi_id
  - total_earnings_lifetime (updated by earnings posting)

---

## Policy & Premium Management (Migration 000004)

### policies
- **structure**: id, worker_id (refs users.id) ✓ Indexed, status, premium_amount, policy_cycle_id
- **status values**: "active", "paused", "cancelled", "skipped", "expired"
- **premium_amount**: Decimal referencing plan tier (12-35 INR range)
- **indexes**: (worker_id, status), policy_cycle_id
- **flow**:
  1. Worker enrolls via EnrollPolicy → creates row with status='active', premium_amount=22
  2. GetPolicy returns policy with dynamic zone & next_due_date
  3. Worker can pause/cancel via PausePolicy/CancelPolicy
  4. Status auto-updates to 'expired' if payment overdue (> 14 days)

### weekly_policy_cycles
- **structure**: id, week_start, week_end, policy_count, total_premium
- **unique**: (week_start, week_end)
- **purpose**: Track policy enrollment windows for billing cohorts

### premium_payments
- **structure**: id, worker_id (refs users.id) ✓ Indexed, policy_id (refs policies.id), amount, status, payment_date
- **status values**: "pending", "completed", "failed"
- **flow**: Worker pays via payPremium → inserts/updates with status='completed' or 'failed'

---

## Earnings & Payment Schedule (Migration 019)

### worker_payments
- **structure**: worker_id (PRIMARY KEY), last_payment_timestamp, next_payment_enabled, coverage_status, created_at, updated_at
- **coverage_status values**: "Active" (< 7 days), "Eligible" (7-14 days), "Expired" (> 14 days from last pay)
- **indexes**: last_payment_timestamp (for schedule queries)
- **flow**:
  1. On delivery completion, getOrBootstrapPaymentSchedule checks this table
  2. Calculates days_since_last_payment and next_payment_enabled
  3. GetPolicy returns payment_status ("Active"/"Eligible"/"Expired")

### weekly_earnings_summary
- **structure**: worker_id (refs users.id), week_start, week_end, total_earnings, claim_eligible, created_at, updated_at
- **unique**: (worker_id, week_start)
- **purpose**: Track weekly earnings for eligibility, claims, and reconciliation
- **updates**: Called in applyWorkerEarningsIncrement after each delivery

---

## Delivery & Batch Management (Migration 000017)

### batches
- **structure**:
  - batch_id (VARCHAR 120, PRIMARY KEY)
  - zone_level (A/B/C), from_city, to_city
  - total_weight, order_count, status
  - pickup_code, delivery_code
  - pickup_user_id (refs users.id) ✓ Indexed, pickup_time, delivery_time
  - batch_earning_inr, earnings_posted (BOOL)
  - created_at, updated_at
- **status values**: "Assigned", "Picked Up", "Delivered"
- **flow**:
  1. **AcceptBatch** (PUT /api/v1/worker/batches/:batch_id/accept):
     - Validates pickup_code using pickupCodeFromBatchID(batchID)
     - Inserts/updates batches row with status='Picked Up'
     - Creates batch_orders entries (child orders)
     - Updates orders table: status='picked_up', worker_id=logged_in_worker
  2. **DeliverBatch** (PUT /api/v1/worker/batches/:batch_id/deliver):
     - **Zone A (same-city)**:
       - Per-order delivery: validate delivery code against deliveryCodeFromOrderID
       - Mark order delivered, immediately call applyWorkerEarningsIncrement
       - Return remainingOrders count
     - **Zone B/C (multi-city)**:
       - Batch-wide delivery: validate code against deliveryCodeFromBatchID
       - Mark all orders delivered
       - Call applyWorkerEarningsIncrement with full batch_earning_inr
     - Guard against double-posting: `if !isZoneA { applyWorkerEarningsIncrement() }`

### batch_orders
- **structure**: (order_id, batch_id) PRIMARY KEY, user_id (refs users.id) ✓ Indexed, status, pickup/delivery_time, delivery_address, contact_name/phone, weight
- **status values**: "Picked Up", "Delivered"
- **purpose**: Denormalization for batch order tracking (linked to orders table)

---

## Claims & Payouts (Migration 000007)

### claims
- **structure**: id, disruption_id (refs disruptions.id) ✓ Indexed, worker_id (refs users.id) ✓ Indexed, claim_amount, status, fraud_verdict, manual_reviewed_at
- **status values**: "pending", "approved", "rejected", "paid"
- **fraud_verdict values**: "approved", "rejected", "flagged_for_review"
- **flow**:
  1. Disruption triggered → ClaimsService.GenerateClaims → creates claims + notifications
  2. Fraud engine (ML) evaluates claim → updates fraud_verdict
  3. Insurer reviews → manual approval or rejection
  4. If approved: PayoutsService creates payout record

### claim_fraud_scores
- **structure**: id, claim_id (UNIQUE refs claims.id), isolation_forest_score, dbscan_score, rule_violations (JSONB), final_verdict
- **purpose**: ML fraud detection scoring
- **linked**: 1-to-1 with claims table

### maintenance_check
- **structure**: id, claim_id (UNIQUE refs claims.id), initiated_date, response_date, findings (TEXT)
- **purpose**: Maintenance check workflow for vehicle damage assessment
- **linked**: 1-to-1 with claims table (optional, if claim requires physical inspection)

### payouts
- **structure**: id, claim_id (UNIQUE refs claims.id), worker_id (refs users.id) ✓ Indexed, amount, status, razorpay_id, razorpay_status
- **status values**: "queued", "processing", "completed", "failed"
- **flow**:
  1. Approved claim → PayoutsService creates row with status='queued'
  2. Batch processor calls Razorpay API → updates razorpay_id, razorpay_status
  3. Payout webhook confirms → status='completed'
  4. PayoutHistoryScreen reads this table sorted by created_at DESC

---

## Disruptions & Notifications (Migrations 000006, 009)

### disruptions
- **structure**: id, zone_id (refs zones.id), disruption_type, severity, description, weather_event, geofence_polygon, created_at
- **flow**: Platform god-mode triggers → creates disruption record + notifications

### disruption_events
- **structure**: event_id (UNIQUE), disruption_id (refs disruptions.id), event_type, event_data (JSONB), recorded_at
- **audit**: Complete event history

### notifications
- **structure**: id, worker_id (refs users.id), notification_type, message, data (JSONB), read_status, created_at
- **types**: "disruption_alert", "claim_made", "payout_completed", "payment_due"
- **read by**: Worker GET /api/v1/worker/notifications (ordered DESC by created_at, limit 50)

---

## Orders & Delivery Tracking (Migration 005+)

### orders
- **structure**: 
  - id, zone_id (refs zones.id)
  - from_city, to_city, from_state, to_state
  - pickup_area, drop_area, address
  - package_weight_kg, tip_inr, order_value
  - customer_name, customer_contact_number
  - worker_id (refs users.id, set on acceptance)
  - status ("assigned", "picked_up", "delivered")
  - accepted_at, picked_up_at, delivered_at
  - created_at, updated_at
- **indexes**: (zone_id), (worker_id), (status)
- **earnings**: Calculated as 60 + tip_inr per delivery

---

## Critical Business Logic

### Earnings Posting (applyWorkerEarningsIncrement)
```sql
-- Updates worker lifetime earnings
UPDATE worker_profiles
SET total_earnings_lifetime = total_earnings_lifetime + amount
WHERE worker_id = ?

-- Weekly aggregation
INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings)
VALUES (?, DATE_TRUNC('week', CURRENT_DATE), ..., amount)
ON CONFLICT (worker_id, week_start) DO UPDATE SET
    total_earnings = total_earnings + amount
```
- **Called**: Per Zone A order delivery OR at Zone B/C batch completion
- **Guard**: `if !isZoneA { applyWorkerEarningsIncrement }` prevents double-posting

### Policy Status Auto-Update
- **GetPolicy** checks payment schedule & auto-updates status to 'expired' if > 14 days

### Payment Schedule Evaluation
```
elapsed < 7 days      → PaymentStatus="Locked", Coverage="Active"
7-14 days             → PaymentStatus="Eligible", Coverage="Active"
> 14 days             → PaymentStatus="Expired", Coverage="Expired"
```

---

## Data Integrity Constraints

| Table | Constraint | Purpose |
|-------|-----------|---------|
| policies | worker_id refs users.id | Orphan prevention |
| batches | pickup_user_id refs users.id | Worker accountability |
| batch_orders | batch_id refs batches.id ON DELETE CASCADE | Cascade cleanup |
| claims | disruption_id, worker_id refs | History linkage |
| payouts | claim_id UNIQUE | Payout per claim |
| worker_payments | worker_id PRIMARY KEY | One schedule per worker |

---

## Migration Checklist

All migrations present and validated:
- ✓ 000001: zones
- ✓ 000002: users (auth_tokens) 
- ✓ 000003: worker_profiles
- ✓ 000004: policies, weekly_policy_cycles, premium_payments
- ✓ 000005: orders (base)
- ✓ 000006: disruptions
- ✓ 000007: claims, claim_fraud_scores, maintenance_check, payouts
- ✓ 000008-018: Order columns, disruption columns, vehicle fields, batch tracking
- ✓ 000019: worker_payments (payment schedule)

**Total: 19 migrations covering 40+ tables**

---

## Deployment Checklist

1. **Database Setup**:
   - PostgreSQL 15+ running
   - `indel_demo` database created
   - Migrations 000001-000019 applied in order
   
2. **Backend Configuration**:
   - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` environment variables set
   - `INDEL_ENV=demo` or `prod`
   - Worker-gateway and platform-gateway configured with DB credentials

3. **Verification**:
   ```sql
   -- Check table counts
   SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';
   -- Expected: ~40 tables after all migrations
   
   -- Verify critical tables
   \dt zones policies claims batches orders worker_profiles
   ```

4. **Operational Flows**:
   - **Delivery Flow**: Order → Batch Accept → Batch Deliver → Earnings Post
   - **Policy Flow**: Enroll → Payment → Coverage Status Update → Auto-Expire
   - **Claims Flow**: Disruption → Claim Creation → Fraud Check → Payout Queue → Razorpay
   - **Notifications**: Worker notifications created on events → Pulled by client

---

## Known Limitations

- Payouts currently show "razorpay_status" but actual Razorpay integration pending
- Maintenance checks not yet integrated into claims workflow
- Disruption geofence polygon not yet used for zone targeting

---

## Version History

- **2026-04-04**: Added dynamic policy data (zone, next_due_date from worker data + payment schedule). Fixed static hardcoded values.
- Fixed delivery earnings per-order posting for Zone A batches.
- All unit tests passing (28 tests across handlers/middleware/services).
