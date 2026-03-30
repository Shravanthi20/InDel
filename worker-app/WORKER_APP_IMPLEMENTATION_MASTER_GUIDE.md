# InDel Worker App Implementation Master Guide

This document is the source of truth for the worker app frontend scope.

---

## What Should Be Built (Frontend Target Scope)

Build the worker app as a delivery-proof plus income-protection product, not as a generic delivery clone.

### Product objective

1. Prove verified work activity.
2. Show income and disruption impact clearly.
3. Trigger transparent claims and payout visibility.
4. Prevent fake activity with fraud-resistant checks.

### Core UX principles

1. No OTP every login for regular users.
2. Orders should be dynamic and empty by default.
3. Delivery flow must include verification checkpoints.
4. Every completed delivery must have customer confirmation.
5. Session inactivity must be detected and tracked.

---

## Final Frontend Feature Scope

### Must build

1. Persistent authentication (register + login).
2. Onboarding and profile setup.
3. Landing page with Start Delivery CTA.
4. Dynamic orders page (empty-first behavior).
5. Fetch verification screen (zone code or IP validation).
6. Delivery execution screen.
7. Delivery completion with customer code verification.
8. Delivery session tracking dashboard.
9. Home dashboard (income + protection summary).
10. Earnings analytics (baseline vs actual vs loss).
11. Policy page with premium and explainability.
12. Claims list and claim detail.
13. Payout history.
14. Notifications center.
15. Profile edit.
16. Debug-only dev tools.

### Nice to have

1. Push notifications with full preferences.
2. Auto inactivity alerts and auto session close.
3. Demo orchestration controls for judges.

### Out of scope

1. Full maps integration.
2. Advanced dispatch and route optimization.
3. Live location streaming UI.

---

## Required Screen Structure

### Authentication and identity

1. Session Gate
2. Register
3. Login
4. Onboarding

### Delivery core flow

5. Landing (Start Delivery)
6. Orders (dynamic, empty-first)
7. Fetch Verification
8. Delivery Execution
9. Delivery Completion
10. Session Tracking

### Income protection visibility

11. Home Dashboard
12. Earnings
13. Policy

### Claims and payouts

14. Claims
15. Claim Detail
16. Payout History

### Account and communication

17. Notifications Center
18. Profile Edit

### Debug/demo

19. Dev Tools (debug only)

---

## Navigation Flow (Target)

1. Session Gate
2. Register or Login
3. Onboarding (first time only)
4. Landing
5. Orders
6. Fetch Verification
7. Delivery Execution
8. Delivery Completion
9. Session Tracking
10. Home, Earnings, Policy, Claims, Payouts, Notifications, Profile
11. Dev Tools (debug only)

---

## Detailed Page Expectations

### 1) Register

Purpose:

- Create persistent worker account.

Fields:

- Email
- Phone
- Username
- Password
- Confirm password

Notes:

- On success, user can be auto-logged in or redirected to login.

### 2) Login

Purpose:

- Fast return login using email or phone plus password.

Fields:

- Email or phone
- Password

Notes:

- OTP should be recovery-only or fallback-only.

### 3) Landing (critical)

Purpose:

- Single clear action after login.

Must include:

- Prominent Start Delivery button.
- Quick snapshot: today deliveries, today earnings, last session.

### 4) Orders (empty-first)

Purpose:

- Show real current work, not static demo cards.

Rules:

1. Initial state should be empty.
2. Show message: No active deliveries. Click Start Delivery or Get More Orders.
3. Populate only via backend or demo-simulation API.

Must include:

- Available deliveries list
- Get More Orders in Zone action
- Accept action per order

### 5) Fetch Verification (anti-fraud gate)

Purpose:

- Verify worker is physically in valid zone before pickup flow continues.

Modes:

1. Zone code entry.
2. Zone/IP validation.

Behavior:

- Without successful verification, delivery cannot move to execution state.

### 6) Delivery Execution

Must show:

- Customer name
- Customer phone
- Address
- Payment type (UPI already paid or pay on delivery)

Actions:

- Confirm and start delivery
- Optional proof photo capture

### 7) Delivery Completion

Purpose:

- Immutable delivery proof via customer code.

Must include:

- Customer code input
- Validate code before marking delivered
- Success state with earning increment

### 8) Session Tracking

Must track:

- Session start/end
- Deliveries completed
- Earnings in session
- Active time
- Fraud signal summary

Inactivity requirement:

1. At 30 minutes idle: show warning.
2. At 60 minutes idle: suggest end session or auto-end (configurable).

### 9) Home Dashboard

Must include:

- Current status
- Earnings snapshot
- Protection status
- Active disruption banner
- Quick links to Orders, Earnings, Policy, Claims

### 10) Earnings

Must visualize:

- Baseline income
- Actual income
- Loss amount
- Potential payout based on coverage ratio

### 11) Policy

Must show:

- Policy status
- Weekly premium
- Coverage ratio
- Next due date
- Explainability chips (risk factors)

### 12) Claims

Must show:

- Claim list with status
- Loss and payout amounts
- Linked disruption context

### 13) Claim Detail

Must show:

- Disruption window
- Loss calculation
- Fraud verdict
- Final payout breakdown

### 14) Payout History

Must show:

- Wallet balance
- Payout timeline
- Claim linkage
- Status and processing timestamp

### 15) Notifications Center

Must support:

- Timeline view
- Read or unread state
- Preferences management
- FCM token registration status

### 16) Profile Edit

Must support:

- Name, zone, vehicle, UPI update
- Read-only identity fields
- Dirty state and validation

### 17) Dev Tools (debug only)

Should include:

- Trigger disruption
- Assign orders
- Simulate deliveries
- Settle earnings
- Reset zone
- Full reset

---

## Frontend Data Contract Requirements

1. Use response wrappers where backend returns envelopes.
2. Do not parse wrapped arrays as plain lists.
3. Keep order lifecycle states explicit:
   - assigned
   - accepted
   - verified
   - picked-up
   - delivered
4. Persist auth token and worker ID in DataStore.
5. Handle 401 by clearing session and redirecting to login.

---

## API Matrix by Screen (Frontend Consumption)

- Register:
  - POST /api/v1/auth/register
- Login:
  - POST /api/v1/auth/login
  - POST /api/v1/auth/otp/send (recovery)
  - POST /api/v1/auth/otp/verify (recovery)
- Onboarding:
  - POST /api/v1/worker/onboard
  - GET /api/v1/worker/profile
- Landing:
  - GET /api/v1/worker/profile
  - GET /api/v1/worker/earnings
  - GET /api/v1/worker/policy
- Orders:
  - GET /api/v1/worker/orders/available
  - GET /api/v1/worker/orders
  - PUT /api/v1/worker/orders/{order_id}/accept
- Fetch Verification:
  - POST /api/v1/worker/fetch-verification/send-code
  - POST /api/v1/worker/fetch-verification/verify
  - GET /api/v1/worker/zone-config
- Delivery Execution:
  - GET /api/v1/worker/orders/{order_id}
  - PUT /api/v1/worker/orders/{order_id}/picked-up
- Delivery Completion:
  - PUT /api/v1/worker/orders/{order_id}/delivered
  - POST /api/v1/worker/orders/{order_id}/code/send
- Session Tracking:
  - GET /api/v1/worker/session/{session_id}
  - GET /api/v1/worker/session/{session_id}/deliveries
  - GET /api/v1/worker/session/{session_id}/fraud-signals
  - PUT /api/v1/worker/session/{session_id}/end
- Home:
  - GET /api/v1/worker/profile
  - GET /api/v1/worker/earnings
  - GET /api/v1/worker/policy
- Earnings:
  - GET /api/v1/worker/earnings
  - GET /api/v1/worker/earnings/history
  - GET /api/v1/worker/earnings/baseline
- Policy:
  - GET /api/v1/worker/policy
  - GET /api/v1/worker/policy/premium
  - POST /api/v1/worker/policy/premium/pay
  - PUT /api/v1/worker/policy/pause
  - PUT /api/v1/worker/policy/cancel
- Claims:
  - GET /api/v1/worker/claims
  - GET /api/v1/worker/claims/{claim_id}
- Payouts:
  - GET /api/v1/worker/wallet
  - GET /api/v1/worker/payouts
- Notifications:
  - GET /api/v1/worker/notifications
  - PUT /api/v1/worker/notifications/preferences
  - POST /api/v1/worker/notifications/fcm-token
- Profile:
  - GET /api/v1/worker/profile
  - PUT /api/v1/worker/profile
- Dev tools:
  - POST /api/v1/demo/trigger-disruption
  - POST /api/v1/demo/assign-orders
  - POST /api/v1/demo/simulate-deliveries
  - POST /api/v1/demo/settle-earnings
  - POST /api/v1/demo/reset-zone
  - POST /api/v1/demo/reset

---

## Build Order (Frontend Execution Plan)

1. Session Gate, Register, Login, token persistence.
2. Onboarding and profile bootstrap.
3. Landing plus Start Delivery path.
4. Orders empty-first plus dynamic fetch.
5. Fetch Verification gate.
6. Delivery Execution and Delivery Completion with customer code.
7. Session Tracking and inactivity handling.
8. Home, Earnings, Policy.
9. Claims, Claim Detail, Payout History.
10. Notifications and Profile Edit.
11. Dev Tools debug surface.

---

## Judge Demo Story (Condensed)

1. Worker registers and logs in.
2. Starts delivery from Landing.
3. Orders appear dynamically after fetch.
4. Worker passes zone verification.
5. Worker completes delivery with customer code.
6. Session tracking shows verified activity.
7. Disruption is triggered.
8. Earnings show baseline versus actual loss.
9. Claim appears with fraud verdict.
10. Payout appears in history and wallet.

---

## What's Already There (Current State)

This section lists current implementation status in the existing worker app and backend integration.

### Already implemented

1. Session gate exists.
2. OTP-based login flow exists.
3. Onboarding exists.
4. Basic home, orders, earnings, policy, claims screens exist.
5. Backend worker gateway is reachable on LAN and health endpoint is working.
6. Worker app base URL configuration has been aligned to worker gateway port 8001.

### Partially implemented

1. Orders lifecycle is partial:
   - Accept action is present.
   - Picked-up and delivered flow integration is incomplete.
2. Claims to payout surfacing is partial:
   - Claims and wallet are present.
   - Full payout timeline rendering is incomplete.
3. Policy detail usage is partial:
   - Premium explainability endpoint is not fully consumed.
4. Earnings details are partial:
   - Baseline endpoint is not fully consumed in dedicated UI flow.

### Not yet implemented in frontend

1. Register plus password login as primary auth.
2. Landing page with Start Delivery primary CTA.
3. Fetch Verification screen.
4. Delivery Execution screen.
5. Delivery Completion with customer code verification.
6. Session Tracking screen.
7. Notifications center full UI.
8. Profile edit full UI flow.
9. Dev Tools screen in app.

### Contract and integration gaps

1. Orders, notifications, and payouts responses need envelope alignment in Retrofit models.
2. FCM token registration API wiring is pending.
3. Notification preferences API wiring is pending.
4. New delivery-first endpoints listed above are still needed for full refined flow.

---

This file intentionally keeps the top focused on what must be built and the bottom focused on what already exists.
