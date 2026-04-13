# InDel Setup Guide

## 1. Required Environment Variables

Create or update the root .env file with at least these values:

HOST_IP=192.168.1.6
API_BASE_URL=http://192.168.1.6:8001/
RAZORPAY_KEY_ID=rzp_test_SZOyjDFEtgTID4
RAZORPAY_KEY_SECRET=Smo4NJ0uCeUO9RWixUtuG5RY

Notes:
- The Android worker app uses RAZORPAY_KEY_ID at build time.
- RAZORPAY_KEY_SECRET is server-side only. Do not ship it in client-side code.

## 2. Start Demo Backend Stack

From repo root:

docker compose -f docker-compose.demo.yml down
docker compose -f docker-compose.demo.yml up -d --build
docker compose -f docker-compose.demo.yml up --force-recreate db-migrate
docker compose -f docker-compose.demo.yml restart worker-gateway

## 3. Build and Install Worker App

From worker-app folder:

.\gradlew.bat clean :app:assembleDebug --no-daemon
.\gradlew.bat :app:installDebug --no-daemon

Important:
- Rebuild after changing .env, because RAZORPAY_KEY_ID is compiled into BuildConfig.

## 4. Seed Orders Continuously for Batch Flow

From scripts folder:

python fake_order_publisher.py --continuous --interval-seconds 60 --orders-per-run 1

This command now publishes zone-aligned orders from:
- zone_a.json
- zone_b.json
- zone_c.json

## 5. Verify Policy Payment Trigger

In the worker app:
1. Open Policy.
2. Tap Touch Pay.
3. Razorpay checkout should open if RAZORPAY_KEY_ID is valid.

If checkout does not open:
- Confirm RAZORPAY_KEY_ID is set in .env.
- Rebuild and reinstall the app.
- Confirm network/API_BASE_URL points to reachable backend host.
