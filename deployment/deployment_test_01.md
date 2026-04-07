# Deployment Test Summary

Date: 2026-04-06
Validation Run: 2026-04-06 03:48:09 +05:30

## Scope

This document records the live deployment validation of the InDel stack across the core backend, the three gateways, and the three ML services. The objective was to verify service availability, route ownership, validation handling, authentication handling, and ML inference readiness.

## Service Endpoints

| Service | Base URL |
|---|---|
| Core | https://indel-backend.onrender.com |
| Worker Gateway | https://indel-worker-gateway.onrender.com |
| Insurer Gateway | https://indel-insurer-gateway.onrender.com |
| Platform Gateway | https://indel-bvvq.onrender.com |
| Premium ML | https://indel-ml-premium.onrender.com |
| Fraud ML | https://indel-ml-fraud.onrender.com |
| Forecast ML | https://indel-ml-forecast.onrender.com |

## Run Summary

| Metric | Value |
|---|---|
| Total endpoints tested | 92 |
| Successful responses (200) | 29 |
| Client/validation responses (4xx) | 51 |
| Server errors (5xx) | 11 |
| No response / timeout | 1 |

## Extended Validation Run

This report also includes the deeper route validation executed after the deployment set was completed and after the database/schema remediation was identified.

| Field | Value |
|---|---|
| Extended run time | 2026-04-06 03:48:09 +05:30 |
| Endpoints tested | 92 |
| Successful responses (200) | 29 |
| Client/validation responses (4xx) | 51 |
| Server errors (5xx) | 11 |
| No response / timeout | 1 |

## Interpretation Rules

| Response Class | Meaning in this report |
|---|---|
| 200 | Request was handled successfully and returned a valid payload |
| 4xx | Expected validation, authentication, or not-found behavior |
| 5xx | Unhandled application/runtime failure and a real defect |

## Service-Level Results

### Core

| Metric | Value |
|---|---|
| Endpoints tested | 9 |
| 200 responses | 1 |
| 4xx responses | 0 |
| 5xx responses | 8 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 13058 | PASS |
| operations | POST | /api/v1/internal/policy/weekly-cycle/run | 500 | 200/4xx/5xx | 510 | FAIL |
| operations | POST | /api/v1/internal/claims/generate-for-disruption/1 | 500 | 200/4xx/5xx | 471 | FAIL |
| operations | POST | /api/v1/internal/claims/auto-process/1 | 500 | 200/4xx/5xx | 553 | FAIL |
| operations | POST | /api/v1/internal/payouts/queue/1 | 500 | 200/4xx/5xx | 613 | FAIL |
| operations | POST | /api/v1/internal/payouts/process | 500 | 200/4xx/5xx | 524 | FAIL |
| operations | GET | /api/v1/internal/payouts/reconciliation | 500 | 200/4xx/5xx | 498 | FAIL |
| operations | POST | /api/v1/internal/data/synthetic/generate | 500 | 200/4xx/5xx | 1217 | FAIL |
| operations | POST | /internal/v1/claims/1/payout | 500 | 200/4xx/5xx | 524 | FAIL |

### Worker Gateway

| Metric | Value |
|---|---|
| Endpoints tested | 55 |
| 200 responses | 10 |
| 4xx responses | 45 |
| 5xx responses | 0 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 479 | PASS |
| health | GET | /api/v1/health | 200 | 200 | 749 | PASS |
| health | GET | /api/v1/status | 200 | 200 | 307 | PASS |
| auth | POST | /api/v1/auth/register | 400 | 400 | 305 | PASS |
| auth | POST | /api/v1/auth/login | 400 | 400 | 306 | PASS |
| auth | POST | /api/v1/auth/otp/send | 400 | 400 | 306 | PASS |
| auth | POST | /api/v1/auth/otp/verify | 400 | 400 | 307 | PASS |
| worker | POST | /api/v1/worker/onboard | 401 | 401/403/4xx | 306 | PASS |
| worker | GET | /api/v1/worker/profile | 401 | 401/403 | 310 | PASS |
| worker | PUT | /api/v1/worker/profile | 401 | 401/403/4xx | 259 | PASS |
| policy | GET | /api/v1/worker/policy | 401 | 401/403 | 357 | PASS |
| policy | POST | /api/v1/worker/policy/enroll | 401 | 401/403/4xx | 302 | PASS |
| policy | PUT | /api/v1/worker/policy/pause | 401 | 401/403/4xx | 311 | PASS |
| policy | PUT | /api/v1/worker/policy/cancel | 401 | 401/403/4xx | 300 | PASS |
| policy | GET | /api/v1/worker/policy/premium | 401 | 401/403 | 305 | PASS |
| policy | POST | /api/v1/worker/policy/premium/pay | 401 | 401/403/4xx | 309 | PASS |
| earnings | GET | /api/v1/worker/earnings | 401 | 401/403 | 306 | PASS |
| earnings | GET | /api/v1/worker/earnings/history | 401 | 401/403 | 308 | PASS |
| earnings | GET | /api/v1/worker/earnings/baseline | 401 | 401/403 | 306 | PASS |
| claims | GET | /api/v1/worker/claims | 401 | 401/403 | 305 | PASS |
| claims | GET | /api/v1/worker/claims/1 | 401 | 401/403/404 | 307 | PASS |
| wallet | GET | /api/v1/worker/wallet | 401 | 401/403 | 250 | PASS |
| payouts | GET | /api/v1/worker/payouts | 401 | 401/403 | 359 | PASS |
| orders | GET | /api/v1/worker/orders | 401 | 401/403 | 417 | PASS |
| orders | GET | /api/v1/worker/orders/available | 200 | 200/401/403 | 505 | PASS |
| orders | GET | /api/v1/worker/orders/assigned | 401 | 401/403 | 305 | PASS |
| orders | GET | /api/v1/worker/orders/1 | 401 | 401/403/404 | 308 | PASS |
| orders | PUT | /api/v1/worker/orders/1/accept | 401 | 401/403/4xx | 409 | PASS |
| orders | PUT | /api/v1/worker/orders/1/picked-up | 401 | 401/403/4xx | 305 | PASS |
| orders | PUT | /api/v1/worker/orders/1/delivered | 401 | 401/403/4xx | 306 | PASS |
| orders | POST | /api/v1/worker/orders/1/code/send | 401 | 401/403/4xx | 309 | PASS |
| verification | POST | /api/v1/worker/fetch-verification/send-code | 401 | 401/403/4xx | 306 | PASS |
| verification | POST | /api/v1/worker/fetch-verification/verify | 401 | 401/403/4xx | 304 | PASS |
| config | GET | /api/v1/worker/zone-config | 401 | 401/403 | 312 | PASS |
| session | GET | /api/v1/worker/session/1 | 401 | 401/403/404 | 301 | PASS |
| session | GET | /api/v1/worker/session/1/deliveries | 401 | 401/403/404 | 307 | PASS |
| session | GET | /api/v1/worker/session/1/fraud-signals | 401 | 401/403/404 | 306 | PASS |
| session | PUT | /api/v1/worker/session/1/end | 401 | 401/403/4xx | 305 | PASS |
| notifications | GET | /api/v1/worker/notifications | 401 | 401/403 | 311 | PASS |
| notifications | PUT | /api/v1/worker/notifications/preferences | 401 | 401/403/4xx | 305 | PASS |
| notifications | POST | /api/v1/worker/notifications/fcm-token | 401 | 401/403/4xx | 265 | PASS |
| demo | POST | /api/v1/demo/trigger-disruption | 401 | 401/403/4xx | 352 | PASS |
| demo | POST | /api/v1/demo/settle-earnings | 401 | 401/403/4xx | 299 | PASS |
| demo | POST | /api/v1/demo/reset-zone | 401 | 401/403/4xx | 307 | PASS |
| demo | POST | /api/v1/demo/reset | 401 | 401/403/4xx | 307 | PASS |
| demo | POST | /api/v1/demo/assign-orders | 401 | 401/403/4xx | 508 | PASS |
| demo | POST | /api/v1/demo/simulate-orders | 401 | 401/403/4xx | 309 | PASS |
| demo | POST | /api/v1/demo/simulate-deliveries | 401 | 401/403/4xx | 310 | PASS |
| demo | POST | /api/v1/demo/orders/publisher/initiate | 200 | 200/401/403/4xx | 303 | PASS |
| demo | POST | /api/v1/demo/orders/publisher/ack | 200 | 200/401/403/4xx | 306 | PASS |
| demo | GET | /api/v1/demo/orders/publisher/status | 200 | 200/401/403 | 305 | PASS |
| demo | POST | /api/v1/demo/orders/ingest | 400 | 200/401/403/4xx | 334 | PASS |
| demo | GET | /api/v1/demo/orders/search | 200 | 200/401/403 | 282 | PASS |
| demo | GET | /api/v1/demo/orders/available | 200 | 200/401/403 | 509 | PASS |
| demo | GET | /api/v1/demo/deliveries | 200 | 200/401/403 | 511 | PASS |

### Insurer Gateway

| Metric | Value |
|---|---|
| Endpoints tested | 11 |
| 200 responses | 7 |
| 4xx responses | 3 |
| 5xx responses | 1 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 1128 | PASS |
| overview | GET | /api/v1/insurer/overview | 200 | 200/4xx | 1433 | PASS |
| overview | GET | /api/v1/insurer/loss-ratio | 200 | 200/4xx | 513 | PASS |
| claims | GET | /api/v1/insurer/claims | 200 | 200/4xx | 715 | PASS |
| claims | GET | /api/v1/insurer/claims/fraud-queue | 200 | 200/4xx | 722 | PASS |
| claims | GET | /api/v1/insurer/claims/1 | 404 | 200/4xx/404 | 502 | PASS |
| claims | POST | /api/v1/insurer/claims/1/review | 400 | 200/4xx/401/403 | 308 | PASS |
| forecast | GET | /api/v1/insurer/forecast | 200 | 200/4xx | 509 | PASS |
| pool | GET | /api/v1/insurer/pool/health | 200 | 200/4xx | 1023 | PASS |
| maintenance | GET | /api/v1/insurer/maintenance-checks | 500 | 200/4xx/401/403 | 718 | FAIL |
| maintenance | POST | /api/v1/insurer/maintenance-checks/1/respond | 400 | 200/4xx/401/403 | 305 | PASS |

### Platform Gateway

| Metric | Value |
|---|---|
| Endpoints tested | 10 |
| 200 responses | 7 |
| 4xx responses | 3 |
| 5xx responses | 0 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 615 | PASS |
| platform | GET | /api/v1/platform/workers | 200 | 200/4xx | 509 | PASS |
| platform | GET | /api/v1/platform/zones | 200 | 200/4xx | 1027 | PASS |
| platform | POST | /api/v1/platform/webhooks/order/assigned | 400 | 200/4xx/5xx | 302 | PASS |
| platform | POST | /api/v1/platform/webhooks/order/completed | 400 | 200/4xx/5xx | 307 | PASS |
| platform | POST | /api/v1/platform/webhooks/order/cancelled | 400 | 200/4xx/5xx | 306 | PASS |
| platform | POST | /api/v1/platform/webhooks/external-signal | 200 | 200/4xx/5xx | 1192 | PASS |
| platform | GET | /api/v1/platform/zones/health | 200 | 200/4xx | 347 | PASS |
| platform | GET | /api/v1/platform/disruptions | 200 | 200/4xx | 508 | PASS |
| demo | POST | /api/v1/demo/trigger-disruption | 200 | 200/4xx/5xx | 1232 | PASS |

### Premium ML

| Metric | Value |
|---|---|
| Endpoints tested | 3 |
| 200 responses | 1 |
| 4xx responses | 0 |
| 5xx responses | 1 |
| No response / timeout | 1 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | timeout | 200 | 60122 | FAIL |
| premium | POST | /ml/v1/premium/calculate | 500 | 200 | 23433 | FAIL |
| premium | POST | /ml/v1/premium/batch-calculate | 500 | 200 | 304 | FAIL |

### Fraud ML

| Metric | Value |
|---|---|
| Endpoints tested | 2 |
| 200 responses | 2 |
| 4xx responses | 0 |
| 5xx responses | 0 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 615 | PASS |
| fraud | POST | /ml/v1/fraud/score | 200 | 200 | 306 | PASS |

### Forecast ML

| Metric | Value |
|---|---|
| Endpoints tested | 2 |
| 200 responses | 2 |
| 4xx responses | 0 |
| 5xx responses | 0 |

| Category | Method | Path | Status | Expected | Latency(ms) | Result |
|---|---|---|---:|---|---:|---|
| health | GET | /health | 200 | 200 | 613 | PASS |
| forecast | POST | /forecast | 200 | 200 | 305 | PASS |

## Root Cause Analysis and Remediation

### Core service failures

The core service failures are caused by a mix of live database/runtime drift and connection-level instability between the deployed Supabase database and the current Go service implementation.

Observed failure patterns:

1. Prepared statement errors such as `stmtcache_* already exists` are still appearing across the core workflow.
2. The synthetic data generation path now fails with `relation "claim_audit_logs" does not exist`, which points to a remaining schema gap in the live database.
3. The earlier missing-column issue has been reduced by the schema remediation, but the live runtime is still not stable enough for the payout and weekly-cycle flows to complete successfully.

Why the SQL change is required:

1. The current backend code expects the production schema to include the audit and payout tracking tables used by the orchestration flows.
2. The production Supabase schema still has gaps relative to those expectations, as shown by the missing `claim_audit_logs` relation.
3. The schema remediation must be paired with a clean connection/pooling configuration so the core runtime paths can execute without repeated 500 errors.

### Insurer gateway failure

The insurer gateway is mostly healthy, but the maintenance-check listing endpoint still returns a 500.

Observed failure pattern:

1. `GET /api/v1/insurer/maintenance-checks` fails with `failed to load maintenance checks`.

Likely cause:

1. The endpoint is still dependent on a live data source or query path that is not fully aligned with the deployed database state.
2. This is an application-level runtime failure, not a route ownership problem.

### Premium ML failures

The premium ML service no longer behaves as a clean health pass on this run: the health probe timed out, and both inference endpoints still return 500.

Observed failure patterns:

1. The health probe exceeded the timeout window, which points to startup or request-handling latency in the service.
2. `POST /ml/v1/premium/calculate` fails during request processing or model prediction.
3. `POST /ml/v1/premium/batch-calculate` fails the same way, which indicates a shared inference-path problem rather than a single-request issue.

Likely cause:

1. The payload encoding and inference preprocessing path in `PremiumModel._preprocess()` and `SHAPExplainer.explain()` is the most likely failure point.
2. The service currently does not wrap prediction/explanation in a defensive error handler, so runtime exceptions are surfaced as 500.

### Why the extended test was added

The extended validation was added after the deployment and schema remediation step so the report could verify:

1. Core route handling after the database schema alignment.
2. Gateway auth and validation behavior across a larger surface area.
3. ML service health and inference readiness.
4. Which remaining 5xx failures were true application defects instead of route mismatches.

## Failures Requiring Follow-Up

| Service | Path | Status | Sample |
|---|---|---:|---|
| core | /api/v1/internal/policy/weekly-cycle/run | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_2\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:00Z"}} |
| core | /api/v1/internal/claims/generate-for-disruption/1 | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_3\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:01Z"}} |
| core | /api/v1/internal/claims/auto-process/1 | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_4\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:01Z"}} |
| core | /api/v1/internal/payouts/queue/1 | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_5\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:02Z"}} |
| core | /api/v1/internal/payouts/process | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_6\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:02Z"}} |
| core | /api/v1/internal/payouts/reconciliation | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_7\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:03Z"}} |
| core | /api/v1/internal/data/synthetic/generate | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: relation \"claim_audit_logs\" does not exist (SQLSTATE 42P01)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:04Z"}} |
| core | /internal/v1/claims/1/payout | 500 | {"error":{"code":"INTERNAL_ERROR","message":"ERROR: prepared statement \"stmtcache_8\" already exists (SQLSTATE 42P05)","details":{}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:13:05Z"}} |
| insurer-gateway | /api/v1/insurer/maintenance-checks | 500 | {"error":{"code":"INTERNAL_ERROR","message":"failed to load maintenance checks","details":{"field":""}},"meta":{"request_id":"req_123","timestamp":"2026-04-05T22:14:35Z"}} |
| premium-ml | /ml/v1/premium/calculate | 500 | Internal Server Error |
| premium-ml | /ml/v1/premium/batch-calculate | 500 | Internal Server Error |

## Supporting Files

| File | Purpose |
|---|---|
| InDel/deployment/run_deployment_deep_validation.ps1 | Consolidated deep endpoint validation runner |
| InDel/deployment/run_deployment_validation.ps1 | Earlier compact validation runner |
| InDel/deployment/run_backend_tests.ps1 | Core service runner |
| InDel/deployment/run_gateway_tests.ps1 | Gateway runner |
| InDel/deployment/run_full_stack_tests.ps1 | Full stack runner |

## Conclusion

The deployed stack is broadly healthy across routing and gateway behavior, but the remaining defects are now narrower and clearer: core runtime paths still fail because of live database/runtime instability, the insurer maintenance listing still returns a 500, and premium ML still has an unstable inference path plus a slow health response.

