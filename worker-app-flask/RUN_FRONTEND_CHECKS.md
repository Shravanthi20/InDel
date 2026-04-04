# Worker App Flask: Run And Frontend Check Guide

This guide shows how to run the Flask worker frontend and verify key routes.

## 1. Prerequisites

- Python environment available (conda/base is fine)
- Backend worker API is running
- Recommended API base: `http://127.0.0.1:8001`
- Alternative via gateway: `http://127.0.0.1:8004` (if your nginx routing is active)

Quick backend route probe:

```powershell
Invoke-WebRequest -Uri http://127.0.0.1:8001/api/v1/auth/login -Method POST -ContentType 'application/json' -Body '{"phone":"9000000000","password":"x"}'
```

A `401`/`400` response is acceptable here. It confirms the route exists.

## 2. Configure Environment

From repository root:

```powershell
Copy-Item worker-app-flask/.env.example worker-app-flask/.env -Force
```

Open `worker-app-flask/.env` and ensure:

```env
API_BASE_URL=http://127.0.0.1:8001
SECRET_KEY=dev-secret-change-me
FLASK_DEBUG=true
```

## 3. Install Dependencies

```powershell
C:/Users/gayat/anaconda3/python.exe -m pip install -r worker-app-flask/requirements.txt
```

## 4. Run Frontend

```powershell
C:/Users/gayat/anaconda3/python.exe -m flask --app worker-app-flask/run.py run --port 5005
```

Open:

- `http://127.0.0.1:5005/login`

## 5. Manual Frontend Verification Checklist

1. Open `/login` and `/register` pages.
2. Register a new user from `/register`.
3. Confirm redirect to `/onboarding`.
4. Submit onboarding form and confirm redirect to `/plan-selection`.
5. Open `/home` and verify dashboard loads.
6. Open `/orders` and verify available/assigned order sections render.
7. Open `/dev-tools` and trigger one demo action (for example, assign orders).

## 6. Expected Behavior

- Unauthenticated access to protected routes (`/home`, `/orders`, `/policy`, `/claims`) redirects to `/login`.
- Successful register stores session token and allows protected routes.
- If backend is unreachable or wrong, register/login remain on page with error message.

## 7. Common Issues

### Register returns 404 page not found

Cause: `API_BASE_URL` points to wrong service (for example `:8003` in your current setup).

Fix: set `API_BASE_URL` to `http://127.0.0.1:8001` (or `:8004` if gateway routing is configured).

### Pages load but data cards are empty

Cause: backend responded but returned no records for the current worker/demo state.

Fix: use `/dev-tools` to seed demo data (assign orders, simulate deliveries, trigger disruption).

### Redirect loop to /login

Cause: session token missing/cleared, or auth call failed.

Fix: re-register/login and verify backend auth routes are reachable.
