# Batch Pickup Flow Test - Sequential Execution

This guide walks through testing the complete batch pickup flow with code validation.

## Prerequisites
- Docker Desktop running
- Python 3.8+ with venv activated
- Browser for HTML simulator
- Android emulator or device for worker-app

---

## Step 1: Start Backend Services

```powershell
cd c:\Users\gayat\projects\get_into\InDel
docker compose -f docker-compose.demo.yml down -v
docker compose -f docker-compose.demo.yml up -d --build
```

**Verify backend is ready:**
```powershell
$resp = Invoke-WebRequest "http://192.168.1.6:8001/api/v1/demo/orders/available?limit=1" -UseBasicParsing
"Backend status: $($resp.StatusCode)"
```

Expected: `200` status code

---

## Step 2: Generate Demo Orders

```powershell
# Ensure venv is activated
(Set-ExecutionPolicy -Scope Process -ExecutionPolicy RemoteSigned) ; (& c:\Users\gayat\projects\get_into\InDel\.venv\Scripts\Activate.ps1)

# Generate orders (use your LAN IP from .env or 192.168.1.6)
python scripts/fake_order_publisher.py --generate --orders-per-run 10 --reset-url http://192.168.1.6:8001/api/v1/demo/reset
```

**Verify batches were created:**
```powershell
$batchesRaw = Invoke-WebRequest "http://192.168.1.6:8001/api/v1/worker/batches?limit=50" -UseBasicParsing
$batches = $batchesRaw.Content | ConvertFrom-Json
Write-Output "Total batches: $($batches.batches.Count)"
Write-Output "First batch: $($batches.batches[0].batchId) with $($batches.batches[0].orderCount) orders"
```

---

## Step 3: Test Pickup Code Validation (Standalone HTML)

**Open in browser:**
```
file:///c:/Users/gayat/projects/get_into/InDel/delivery_batch_pickup_simulator.html
```

**Test flow:**
1. Toggle "Show pickup codes" to see the 4-digit codes
2. Click any batch card to open details
3. Enter the pickup code shown in the modal
4. Verify "Batch accepted" message appears
5. Verify batch status changes to "Picked Up"

This validates the local code generation and validation logic before hitting the real app.

---

## Step 4: Build and Run Worker App

**Build the Android app:**
```powershell
cd c:\Users\gayat\projects\get_into\InDel\worker-app
# Gradle build with Android Studio or:
./gradlew build
```

**Run on emulator/device:**
```powershell
# Launch emulator first or have device connected
./gradlew installDebug
./gradlew assembleDebug
```

**Or directly from Android Studio:**
- Open `worker-app/` as a project
- Configure Android emulator or connect physical device
- Set LAN IP in `worker-app/.env` (copy from root `.env`)
- Run → Run 'app'

---

## Step 5: Test Real App Batch Pickup Flow

1. **Login** with demo credentials:
   - Phone: +919999999999
   - OTP: 123456

2. **Navigate to Orders tab** → See "Available Near You" batches

3. **Tap a batch card** → Opens batch detail screen

4. **Tap "Accept Batch"** → Triggers pickup code dialog

5. **Enter the 4-digit code** shown on the batch card (randomly generated per batch, visible only in app)

6. **Submit** → Backend accepts all orders in batch:
   - Status changes to "Accepted"
   - Orders now appear under "Active Tasks" on refresh
   - Backend has assigned orders to this worker

---

## Verification Checklist

- [ ] Backend containers running (check `docker compose -f docker-compose.demo.yml ps`)
- [ ] Orders created via publisher script
- [ ] HTML simulator shows batches and codes
- [ ] Pickup code validation works in simulator
- [ ] Worker app connects to backend (see diagnostics in Orders tab)
- [ ] Batch detail shows correct order count and weight
- [ ] Code acceptance updates backend (orders move to assigned)
- [ ] Status changes are reflected on next refresh

---

## Troubleshooting

**Backend not responding:**
```powershell
docker compose -f docker-compose.demo.yml logs backend
```

**No batches generated:**
- Verify orders were created: check demo/orders/available endpoint
- Check publisher output for errors
- Ensure reset-url is correct (matches LAN IP config)

**App can't connect to backend:**
- Verify `.env` in worker-app has correct `API_BASE_URL`
- Check LAN IP is accessible: `ping 192.168.1.6`
- Verify firewall allows port 8001 and 8003

**Pickup code mismatch:**
- Code is per-batch and generated on app startup
- Verify you're entering the exact 4-digit code shown
- Clear app cache and restart if issues persist

---

## Notes

- The pickup code is generated **per batch** in the batch detail screen and stored in Compose state
- Code validation happens **locally** in the app before calling backend
- Backend `AcceptBatch` endpoint (`POST /api/v1/worker/batches/:batch_id/accept`) updates all orders in the batch to `accepted` status
- Orders then appear in "Active Tasks" section on next fetch
