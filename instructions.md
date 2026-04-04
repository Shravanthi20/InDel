# Backend Implementation Instructions (InDel)

This document is the single onboarding and execution guide for implementing backend features in InDel.

Goal: if a new developer reads only this file, they should be able to set up the backend, understand the structure, add a feature safely, and verify it end-to-end.

## 1. What You Need

- Windows, macOS, or Linux
- Docker Desktop with Docker Compose v2
- Go 1.25+
- PostgreSQL client tools (optional but useful)
- Git

Optional but useful:
- VS Code
- Postman or curl

## 2. High-Level System

InDel backend is split into gateways plus core:

- core service: domain logic and internal orchestration
- worker-gateway: worker-facing APIs (auth, orders, policy, claims, payouts, notifications)
- platform-gateway: platform ops and disruption APIs
- insurer-gateway: insurer dashboards and review APIs
- api-gateway (nginx): unified entrypoint (default localhost:8004)

ML services run separately and integrate with backend:
- premium-ml
- fraud-ml
- forecast-ml

Primary persistence:
- PostgreSQL

## 3. Repo Map You Will Use Most

- backend/cmd/core
- backend/cmd/worker-gateway
- backend/cmd/platform-gateway
- backend/cmd/insurer-gateway
- backend/internal/handlers
- backend/internal/router
- backend/internal/services
- backend/internal/models
- migrations
- docker-compose.demo.yml
- docs/API_DESIGN.md
- docs/DATABASE_DESIGN.md
- docs/RUNNING.md

## 4. Quick Start (Recommended)

From repository root:

1. Clean start
- docker compose -f docker-compose.demo.yml down -v

2. Build and run everything
- docker compose -f docker-compose.demo.yml up -d --build

3. Verify health
- http://localhost:8000/health (core)
- http://localhost:8001/health (worker-gateway)
- http://localhost:8002/health (insurer-gateway)
- http://localhost:8003/health (platform-gateway)
- http://localhost:8004/health (api-gateway)

Unified API entrypoint for most manual tests:
- http://localhost:8004

## 5. Running Tests

From backend folder:

- go test ./...

For targeted handler checks during feature work:
- go test ./internal/handlers/worker -v
- go test ./internal/handlers/platform -v

## 6. How Routing Is Organized

Routers live under backend/internal/router.

Typical mapping:
- Worker endpoints are registered in worker_router.go
- Platform endpoints are registered in platform router files
- Insurer endpoints are registered in insurer router files

When adding a new endpoint:
1. Add handler function in the relevant domain package under backend/internal/handlers
2. Register route in the matching router
3. Ensure middleware/auth expectations are consistent with existing endpoints
4. Add tests for happy path and key failure paths

## 7. Database and Migrations Rules

All schema changes must be migration-driven.

Location:
- migrations

Rules:
1. Create both up and down migration files
2. Use a new unique version number (never reuse existing versions)
3. Keep migration changes focused and reversible
4. Add indexes for new query-heavy columns
5. If adding ON CONFLICT logic, ensure unique constraints match the conflict target exactly

Common migration pitfalls already seen in this project:
- Duplicate migration version numbers break migrate startup
- Upsert statements fail if expected unique constraints are missing
- Handler SQL that updates updated_at fails if column is missing

## 8. Feature Implementation Workflow (Recommended)

Use this exact flow each time:

1. Define behavior clearly
- Input payload
- Output payload
- Error codes and error strings
- DB side effects

2. Check existing patterns first
- Similar handlers in same package
- Existing model structs and response shapes
- Existing status enums

3. Implement handler logic
- Validate request body early
- Check auth/ownership constraints
- Use transactions when multiple writes must be atomic
- Return stable, explicit error identifiers

4. Update routes
- Register endpoint in the proper router file

5. Add/update tests
- Happy path
- Invalid input
- Not found/unauthorized
- Idempotency or state-transition guard

6. Run tests
- Start with targeted package, then full backend tests

7. Verify live with Docker stack
- Trigger endpoint through localhost:8004
- Confirm DB write/read behavior with psql if needed

## 9. Error Handling Conventions

Follow the existing backend style:
- 400 for invalid request/input/code errors
- 404 for not found/not assignable
- 409 for state conflicts (already delivered, not picked up, etc.)
- 500 for unexpected internal failures

Prefer stable machine-readable error values in JSON, for example:
- incorrect_delivery_code
- batch_not_picked_up
- batch_not_found_or_not_assignable

Do not replace precise errors with generic text if client UX depends on the error key.

## 10. Status Transition Discipline

For delivery/policy/claims workflows, protect transitions carefully.

Examples:
- Do not allow delivery on assigned batch if pickup has not happened
- Do not re-credit earnings for already posted states
- Do not generate duplicate claims for same disruption+worker pair

If transition logic depends on current DB state, re-read inside the transaction.

## 11. End-to-End Validation Checklist

Before considering backend work complete:

1. Handler tests pass
2. Full backend tests pass
3. Docker services rebuilt and healthy
4. Endpoint works through api-gateway
5. Data write verified in DB
6. Data fetch verified in API response
7. Regression check done for adjacent flows

## 12. API and DB Documentation Sources of Truth

- API contract guidance: docs/API_DESIGN.md
- Database reference: docs/DATABASE_DESIGN.md
- Runtime commands: docs/RUNNING.md

If your implementation intentionally differs from docs, update docs in the same PR.

## 13. Practical Troubleshooting

If endpoint returns 502 via gateway:
- Check corresponding gateway/core container is running
- Check logs for startup failures

If route appears wired but returns empty data:
- Verify DB env vars inside the correct gateway container
- Verify the expected database is being used

If claims do not appear after disruption trigger:
- Confirm disruption was written
- Confirm final payout was greater than zero
- Confirm workers exist in affected zone
- Confirm no duplicate-block condition is preventing inserts

If delivery error behavior seems wrong:
- Re-check order of state checks in handler
- Ensure state guard is not shadowed by earlier ownership/not-found checks

## 14. Definition of Done for Backend Features

A backend feature is done only when:
- Route exists and is reachable
- Handler logic works for happy path and key failures
- DB schema supports the behavior safely
- Tests cover behavior changes
- Docker/live verification is completed
- Documentation is updated if behavior changed

## 15. Suggested PR Structure

Keep PRs reviewable:
- Commit 1: migrations
- Commit 2: handler/service logic
- Commit 3: router wiring
- Commit 4: tests
- Commit 5: docs and verification notes

This helps future contributors understand intent and rollback impact quickly.
