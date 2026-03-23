#!/usr/bin/env python3
"""
Mock backend for InDel worker Android app.

Run:
  python worker-app/mock-backend/mock_worker_backend.py
"""

from __future__ import annotations

import json
import os
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any
from urllib.parse import parse_qs, urlparse

HOST = os.environ.get("INDEL_MOCK_HOST", "0.0.0.0")
PORT = int(os.environ.get("INDEL_MOCK_PORT", "8082"))


class MockState:
    def __init__(self) -> None:
        self.phone_to_otp = {"+919999999999": "123456"}
        self.token_to_worker_id = {"mock-jwt-token": "worker-001"}
        self.worker_profiles = {
            "worker-001": {
                "worker_id": "worker-001",
                "name": "Gayathri Worker",
                "phone": "+919999999999",
                "zone": "Tambaram, Chennai",
                "vehicle_type": "bike",
                "upi_id": "gayathri@upi",
                "coverage_status": "active",
                "enrolled": True,
            }
        }
        self.policy = {
            "policy_id": "pol-001",
            "status": "active",
            "weekly_premium_inr": 22,
            "coverage_ratio": 0.8,
            "zone": "Tambaram, Chennai",
            "next_due_date": "2026-03-30",
            "shap_breakdown": [
                {"feature": "rain_risk", "impact": 0.42},
                {"feature": "order_drop_volatility", "impact": 0.31},
                {"feature": "historical_disruptions", "impact": 0.27},
            ],
        }
        self.earnings = {
            "currency": "INR",
            "this_week_actual": 3120,
            "this_week_baseline": 4080,
            "protected_income": 3264,
            "history": [
                {"week": "2026-W08", "actual": 3520, "baseline": 3980},
                {"week": "2026-W09", "actual": 3410, "baseline": 4010},
                {"week": "2026-W10", "actual": 3290, "baseline": 4050},
                {"week": "2026-W11", "actual": 3120, "baseline": 4080},
            ],
        }
        self.claims = [
            {
                "claim_id": "clm-001",
                "status": "approved",
                "zone": "Tambaram, Chennai",
                "disruption_type": "heavy_rain",
                "disruption_window": {
                    "start": "2026-03-18T11:00:00Z",
                    "end": "2026-03-18T16:00:00Z",
                },
                "income_loss": 870,
                "payout_amount": 696,
                "fraud_verdict": "clear",
                "created_at": "2026-03-18T16:20:00Z",
            }
        ]
        self.wallet = {
            "currency": "INR",
            "available_balance": 1580,
            "last_payout_amount": 696,
            "last_payout_at": "2026-03-19T09:10:00Z",
        }
        self.payouts = [
            {
                "payout_id": "pay-001",
                "claim_id": "clm-001",
                "amount": 696,
                "method": "upi",
                "status": "processed",
                "processed_at": "2026-03-19T09:10:00Z",
            }
        ]


STATE = MockState()


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")


class Handler(BaseHTTPRequestHandler):
    server_version = "InDelMock/1.0"

    def do_OPTIONS(self) -> None:  # noqa: N802
        self.send_response(204)
        self._send_cors_headers()
        self.end_headers()

    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        path = parsed.path
        query = parse_qs(parsed.query)

        if path == "/api/v1/health":
            return self._ok({"ok": True, "service": "worker-mock", "time": now_iso()})
        if path == "/api/v1/status":
            return self._ok({"status": "up", "environment": "mock", "time": now_iso()})

        worker_id = self._auth_worker_id()
        if worker_id is None:
            return

        if path == "/api/v1/worker/profile":
            return self._ok({"worker": STATE.worker_profiles.get(worker_id)})
        if path == "/api/v1/worker/policy":
            return self._ok({"policy": STATE.policy})
        if path == "/api/v1/worker/policy/premium":
            payload = {
                "weekly_premium_inr": STATE.policy["weekly_premium_inr"],
                "currency": "INR",
                "shap_breakdown": STATE.policy["shap_breakdown"],
            }
            return self._ok(payload)
        if path == "/api/v1/worker/earnings":
            return self._ok(STATE.earnings)
        if path == "/api/v1/worker/earnings/history":
            return self._ok({"history": STATE.earnings["history"]})
        if path == "/api/v1/worker/earnings/baseline":
            return self._ok({"baseline": STATE.earnings["this_week_baseline"], "currency": "INR"})
        if path == "/api/v1/worker/claims":
            return self._ok({"claims": STATE.claims})
        if path.startswith("/api/v1/worker/claims/"):
            claim_id = path.rsplit("/", 1)[-1]
            match = next((claim for claim in STATE.claims if claim["claim_id"] == claim_id), None)
            if not match:
                return self._not_found("claim_not_found")
            return self._ok(match)
        if path == "/api/v1/worker/wallet":
            return self._ok(STATE.wallet)
        if path == "/api/v1/worker/payouts":
            limit = int(query.get("limit", ["10"])[0])
            return self._ok({"payouts": STATE.payouts[: max(limit, 1)]})

        return self._not_found("endpoint_not_found")

    def do_POST(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        path = parsed.path
        body = self._read_json_body()

        if path == "/api/v1/auth/otp/send":
            phone = body.get("phone")
            if not phone:
                return self._bad_request("phone_required")
            STATE.phone_to_otp[phone] = "123456"
            return self._ok(
                {
                    "message": "otp_sent",
                    "otp_for_testing": "123456",
                    "phone": phone,
                    "expires_in_seconds": 300,
                }
            )

        if path == "/api/v1/auth/otp/verify":
            phone = body.get("phone")
            otp = str(body.get("otp", ""))
            expected_otp = STATE.phone_to_otp.get(phone)
            if not expected_otp or expected_otp != otp:
                return self._unauthorized("invalid_otp")

            token = "mock-jwt-token"
            worker_id = "worker-001"
            STATE.token_to_worker_id[token] = worker_id
            if worker_id not in STATE.worker_profiles:
                STATE.worker_profiles[worker_id] = {
                    "worker_id": worker_id,
                    "name": "New Worker",
                    "phone": phone,
                    "zone": "Tambaram, Chennai",
                    "vehicle_type": "bike",
                    "upi_id": "new@upi",
                    "coverage_status": "inactive",
                    "enrolled": False,
                }

            return self._ok(
                {
                    "message": "otp_verified",
                    "token": token,
                    "token_type": "Bearer",
                    "worker_id": worker_id,
                }
            )

        worker_id = self._auth_worker_id()
        if worker_id is None:
            return

        if path == "/api/v1/worker/onboard":
            profile = STATE.worker_profiles.get(worker_id, {})
            profile["name"] = body.get("name", profile.get("name", "New Worker"))
            profile["zone"] = body.get("zone", profile.get("zone", "Tambaram, Chennai"))
            profile["vehicle_type"] = body.get("vehicle_type", profile.get("vehicle_type", "bike"))
            profile["upi_id"] = body.get("upi_id", profile.get("upi_id", "new@upi"))
            STATE.worker_profiles[worker_id] = profile
            return self._ok({"message": "onboarded", "worker": profile})

        if path == "/api/v1/worker/policy/enroll":
            STATE.policy["status"] = "active"
            worker = STATE.worker_profiles.get(worker_id)
            if worker:
                worker["enrolled"] = True
                worker["coverage_status"] = "active"
            return self._ok({"message": "policy_enrolled", "policy": STATE.policy})

        if path == "/api/v1/worker/policy/premium/pay":
            amount = body.get("amount", STATE.policy["weekly_premium_inr"])
            return self._ok(
                {
                    "message": "payment_successful",
                    "amount": amount,
                    "currency": "INR",
                    "payment_id": "mock-payment-001",
                }
            )

        return self._not_found("endpoint_not_found")

    def do_PUT(self) -> None:  # noqa: N802
        path = urlparse(self.path).path

        worker_id = self._auth_worker_id()
        if worker_id is None:
            return

        if path == "/api/v1/worker/policy/pause":
            STATE.policy["status"] = "paused"
            worker = STATE.worker_profiles.get(worker_id)
            if worker:
                worker["coverage_status"] = "paused"
            return self._ok({"message": "policy_paused", "policy": STATE.policy})

        if path == "/api/v1/worker/policy/cancel":
            STATE.policy["status"] = "cancelled"
            worker = STATE.worker_profiles.get(worker_id)
            if worker:
                worker["coverage_status"] = "inactive"
                worker["enrolled"] = False
            return self._ok({"message": "policy_cancelled", "policy": STATE.policy})

        return self._not_found("endpoint_not_found")

    def log_message(self, format: str, *args: Any) -> None:  # noqa: A003
        # Keep logs concise while still showing useful request info.
        print(f"[{self.log_date_time_string()}] {self.address_string()} {format % args}")

    def _auth_worker_id(self) -> str | None:
        auth = self.headers.get("Authorization", "")
        if not auth.startswith("Bearer "):
            self._unauthorized("missing_or_invalid_bearer_token")
            return None
        token = auth.replace("Bearer ", "", 1).strip()
        worker_id = STATE.token_to_worker_id.get(token)
        if not worker_id:
            self._unauthorized("unknown_token")
            return None
        return worker_id

    def _read_json_body(self) -> dict[str, Any]:
        raw_length = self.headers.get("Content-Length", "0")
        try:
            length = int(raw_length)
        except ValueError:
            return {}
        if length <= 0:
            return {}
        raw = self.rfile.read(length)
        try:
            parsed = json.loads(raw.decode("utf-8"))
            return parsed if isinstance(parsed, dict) else {}
        except json.JSONDecodeError:
            return {}

    def _send_json(self, status: int, payload: dict[str, Any]) -> None:
        body = json.dumps(payload, ensure_ascii=True).encode("utf-8")
        self.send_response(status)
        self._send_cors_headers()
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _send_cors_headers(self) -> None:
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

    def _ok(self, payload: dict[str, Any]) -> None:
        self._send_json(200, payload)

    def _bad_request(self, error: str) -> None:
        self._send_json(400, {"error": error})

    def _unauthorized(self, error: str) -> None:
        self._send_json(401, {"error": error})

    def _not_found(self, error: str) -> None:
        self._send_json(404, {"error": error})


def run() -> None:
    server = ThreadingHTTPServer((HOST, PORT), Handler)
    print(f"InDel worker mock backend running at http://{HOST}:{PORT}")
    print("Press Ctrl+C to stop")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()


if __name__ == "__main__":
    run()
