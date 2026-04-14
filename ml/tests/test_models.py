"""
Model integration tests — hits the live Docker containers directly.
Tests all 3 ML services: forecast, fraud, premium.

Run from the ml/ directory:
    python tests/test_models.py

Or with pytest:
    pytest tests/test_models.py -v
"""
import sys
import json
import urllib.request
import urllib.error
from datetime import date

# ── Helpers ───────────────────────────────────────────────────────────────────

def post(url: str, body: dict) -> dict:
    data = json.dumps(body).encode()
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read())

def get(url: str) -> dict:
    with urllib.request.urlopen(url, timeout=10) as resp:
        return json.loads(resp.read())

PASS = "\033[92m✅ PASS\033[0m"
FAIL = "\033[91m❌ FAIL\033[0m"
results = []

def check(name: str, condition: bool, detail: str = ""):
    status = PASS if condition else FAIL
    print(f"  {status}  {name}" + (f"  [{detail}]" if detail else ""))
    results.append(condition)
    return condition

# ── SECTION 1: Forecast ML (port 9003) ───────────────────────────────────────

print("\n\033[96m━━━ FORECAST ML  (localhost:9003) ━━━\033[0m")

try:
    h = get("http://localhost:9003/health")
    check("Health endpoint responds", h.get("status") == "ok")
    check("All 4 zones trained (no fallback)", h.get("zones_fallback") == [], str(h.get("zones_fallback")))
    check("zones_trained contains [1,2,3,4]", sorted(h.get("zones_trained", [])) == [1, 2, 3, 4])

    # Zone 1 forecast
    r = post("http://localhost:9003/forecast", {"zone_id": 1})
    check("Forecast returns 7 days", len(r.get("forecast", [])) == 7)
    check("Inference mode = seasonal (data-driven)", r.get("inference") == "seasonal", r.get("inference"))
    check("Dates are from today onwards", r["forecast"][0]["date"] == date.today().isoformat(), r["forecast"][0]["date"])
    probs = [f["disruption_probability"] for f in r["forecast"]]
    check("All probabilities in [0,1]", all(0 <= p <= 1 for p in probs), str(probs))
    check("Purpose states reserve planning only", "reserve planning" in r.get("purpose", "").lower())

    # Unknown zone → 404
    try:
        post("http://localhost:9003/forecast", {"zone_id": 99})
        check("Unknown zone returns error", False, "Expected 404, got 200")
    except urllib.error.HTTPError as e:
        check("Unknown zone returns 404", e.code == 404, f"HTTP {e.code}")

    # Model info endpoint
    info = get("http://localhost:9003/model-info")
    check("Model info has known_limitations", isinstance(info.get("known_limitations"), list))
    check("Model info states reserve planning purpose", "reserve planning" in info.get("purpose", "").lower())
    check("Retraining cadence documented", "Weekly" in info.get("retraining_cadence", ""))
    check("DeepAR upgrade path documented", "DeepAR" in info.get("upgrade_path", ""))

    # Zones endpoint
    zones = get("http://localhost:9003/zones")
    check("Zones endpoint lists 4 zones", len(zones.get("zones", [])) == 4)

    # Zone 2, 3, 4 spot-check
    for zid in [2, 3, 4]:
        r2 = post("http://localhost:9003/forecast", {"zone_id": zid})
        check(f"Zone {zid} forecast returns 7 days", len(r2.get("forecast", [])) == 7)

except Exception as e:
    print(f"  \033[91m❌ FORECAST service unreachable: {e}\033[0m")

# ── SECTION 2: Fraud ML (port 9002) ──────────────────────────────────────────

print("\n\033[96m━━━ FRAUD ML  (localhost:9002) ━━━\033[0m")

FRAUD_SAMPLE = {
    "claim_id": 9999,
    "worker_id": 1,
    "zone_id": 1,
    "claim_amount": 425.0,
    "baseline_earnings": 5000.0,
    "disruption_type": "demand_drop",
    "gps_in_zone": True,
    "deliveries_during_disruption": 2,
    "zone_avg_claim_amount": 400.0,
    "worker_history": {
        "total_claims_last_8_weeks": 0,
        "avg_claim_amount": 0.0,
        "earnings_variance": 0.1,
        "zone_change_count": 0,
        "days_active": 180,
        "delivery_attempt_rate": 0.85,
    },
}

FRAUD_URL = "http://localhost:9002/ml/v1/fraud/score"
try:
    h = get("http://localhost:9002/health")
    check("Health endpoint responds", h.get("status") == "ok")
    check("3 fraud layers reported", len(h.get("layers", [])) == 3, str(h.get("layers")))

    r = post(FRAUD_URL, FRAUD_SAMPLE)
    check("Score returns fraud_score float", isinstance(r.get("fraud_score"), (int, float)))
    check("Score in [0,1]", 0 <= r.get("fraud_score", -1) <= 1, str(r.get("fraud_score")))
    check("verdict is clear/review/flagged/auto_reject", r.get("verdict") in ("clear", "review", "flagged", "auto_reject"), r.get("verdict"))
    check("signals is a list", isinstance(r.get("signals"), list))
    check("routing field present", r.get("routing") in ("auto_approve", "manual_review", "auto_reject"), r.get("routing"))

    # High-risk claim
    risky = {**FRAUD_SAMPLE, "claim_amount": 999999.0}
    risky["worker_history"] = {**FRAUD_SAMPLE["worker_history"], "total_claims_last_8_weeks": 20}
    r2 = post(FRAUD_URL, risky)
    check("High-risk claim flagged (score > 0.3)", r2.get("fraud_score", 0) > 0.3, str(r2.get("fraud_score")))

except Exception as e:
    print(f"  \033[91m❌ FRAUD service error: {e}\033[0m")

# ── SECTION 3: Premium ML (port 9001) ────────────────────────────────────────

print("\n\033[96m━━━ PREMIUM ML  (localhost:9001) ━━━\033[0m")

PREMIUM_SAMPLE = {
    "worker_id": "wkr_test_001",
    "zone_id": "zone_chennai_coastal",
    "city": "Chennai",
    "state": "Tamil Nadu",
    "zone_type": "coastal",
    "vehicle_type": "two_wheeler",
    "season": "Monsoon",
    "experience_days": 400,
    "avg_daily_orders": 18.0,
    "avg_daily_earnings": 1100.0,
    "active_hours_per_day": 8.0,
    "rainfall_mm": 45.0,
    "aqi": 70.0,
    "temperature": 30.0,
    "humidity": 80.0,
    "order_volatility": 0.18,
    "earnings_volatility": 0.20,
    "recent_disruption_rate": 0.08,
}

try:
    h = get("http://localhost:9001/health")
    check("Health endpoint responds", h.get("status") == "ok")
    check("Model loaded flag is true", h.get("model_loaded") == True)

    r = post("http://localhost:9001/ml/v1/premium/calculate", PREMIUM_SAMPLE)
    d = r.get("data", {})
    check("Premium amount is positive", d.get("premium_inr", 0) > 0, str(d.get("premium_inr")))
    check("Risk score in [0,1]", 0 <= d.get("risk_score", -1) <= 1, str(d.get("risk_score")))
    check("Explainability provided", isinstance(d.get("explainability"), list) and len(d.get("explainability")) > 0)
    check("Model version present", bool(d.get("model_version")))

    # High-risk worker should get higher premium
    risky_worker = {**PREMIUM_SAMPLE, "rainfall_mm": 200.0, "aqi": 180.0, "recent_disruption_rate": 0.5}
    r2 = post("http://localhost:9001/ml/v1/premium/calculate", risky_worker)
    d2 = r2.get("data", {})
    check("Risky worker gets higher premium", d2.get("premium_inr", 0) >= d.get("premium_inr", 0),
          f"{d2.get('premium_inr')} vs {d.get('premium_inr')}")
    check("Risky worker has higher risk score", d2.get("risk_score", 0) >= d.get("risk_score", 0),
          f"{d2.get('risk_score')} vs {d.get('risk_score')}")

except Exception as e:
    print(f"  \033[91m❌ PREMIUM service error: {e}\033[0m")

# ── Summary ───────────────────────────────────────────────────────────────────

passed = sum(results)
total = len(results)
pct = int(passed / total * 100) if total else 0
color = "\033[92m" if pct == 100 else "\033[93m" if pct >= 70 else "\033[91m"
print(f"\n{color}━━━ RESULTS: {passed}/{total} passed ({pct}%) ━━━\033[0m\n")
sys.exit(0 if pct == 100 else 1)
