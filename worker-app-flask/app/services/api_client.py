"""
API Client â€” mirrors Kotlin's WorkerApiService + AuthInterceptor + NetworkModule.

All backend calls go through this module.
JWT token is automatically injected from Flask session (mirrors PreferencesDataStore).
401 responses â†’ clear session and signal redirect to login.
"""
import requests
from flask import current_app, session as flask_session


class ApiError(Exception):
    """Raised when the backend returns an error status."""
    def __init__(self, status_code: int, message: str):
        self.status_code = status_code
        self.message = message
        super().__init__(message)


def _base_url() -> str:
    return current_app.config["API_BASE_URL"]


def _headers() -> dict:
    """Inject JWT token from session â€” mirrors AuthInterceptor."""
    headers = {"Content-Type": "application/json"}
    token = flask_session.get("token")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def _request(method: str, url: str, **kwargs):
    timeout_seconds = current_app.config.get("API_TIMEOUT_SECONDS", 12)
    kwargs.setdefault("timeout", timeout_seconds)
    return requests.request(method=method, url=url, **kwargs)


def _handle_response(resp: requests.Response):
    """
    Returns parsed JSON on success.
    Clears session and raises ApiError on 401.
    Raises ApiError on other non-2xx responses.
    """
    if resp.status_code == 401:
        flask_session.clear()
        raise ApiError(401, "Unauthorized â€” please log in again.")
    if not resp.ok:
        try:
            detail = resp.json().get("message") or resp.json().get("error") or resp.text
        except Exception:
            detail = resp.text
        raise ApiError(resp.status_code, detail)
    try:
        return resp.json()
    except Exception:
        return {}


# â”€â”€ Auth â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def register(username: str, phone: str, email: str, password: str,
             zone_level: str = None, zone_name: str = None):
    payload = {"username": username, "phone": phone, "email": email,
               "password": password}
    if zone_level:
        payload["zone_level"] = zone_level
    if zone_name:
        payload["zone_name"] = zone_name
    resp = _request("POST", f"{_base_url()}/api/v1/auth/register",
                         json=payload, headers=_headers())
    return _handle_response(resp)


def login(password: str, email: str = None, phone: str = None):
    payload = {"password": password}
    if email:
        payload["email"] = email
    if phone:
        payload["phone"] = phone
    resp = _request("POST", f"{_base_url()}/api/v1/auth/login",
                         json=payload, headers=_headers())
    return _handle_response(resp)


def send_otp(phone: str):
    resp = _request("POST", f"{_base_url()}/api/v1/auth/otp/send",
                         json={"phone": phone}, headers=_headers())
    return _handle_response(resp)


def verify_otp(phone: str, otp: str):
    resp = _request("POST", f"{_base_url()}/api/v1/auth/otp/verify",
                         json={"phone": phone, "otp": otp}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Worker Profile â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_profile():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/profile", headers=_headers())
    return _handle_response(resp)


def onboard(name: str, vehicle_type: str, upi_id: str,
            zone_level: str = None, zone_name: str = None,
            zone_id: int = None, city: str = None,
            from_city: str = None, to_city: str = None,
            vehicle_name: str = None):
    payload = {"name": name, "vehicle_type": vehicle_type, "upi_id": upi_id}
    if zone_level:
        payload["zone_level"] = zone_level
    if zone_name:
        payload["zone_name"] = zone_name
    if zone_id is not None:
        payload["zone_id"] = zone_id
    if city:
        payload["city"] = city
    if from_city:
        payload["from_city"] = from_city
    if to_city:
        payload["to_city"] = to_city
    if vehicle_name:
        payload["vehicle_name"] = vehicle_name
    resp = _request("POST", f"{_base_url()}/api/v1/worker/onboard",
                         json=payload, headers=_headers())
    return _handle_response(resp)


def update_profile(name: str, zone_level: str, zone_name: str,
                   vehicle_type: str, upi_id: str, zone_id: int = None,
                   city: str = None, from_city: str = None, to_city: str = None):
    payload = {"name": name, "zone_level": zone_level, "zone_name": zone_name,
               "vehicle_type": vehicle_type, "upi_id": upi_id}
    if zone_id is not None:
        payload["zone_id"] = zone_id
    if city:
        payload["city"] = city
    if from_city:
        payload["from_city"] = from_city
    if to_city:
        payload["to_city"] = to_city
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/profile",
                        json=payload, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Zones â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_zones():
    resp = _request("GET", f"{_base_url()}/api/v1/platform/zones", headers=_headers())
    return _handle_response(resp)


def get_zone_config():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/zone-config", headers=_headers())
    return _handle_response(resp)


def get_zone_paths(path_type: str):
    resp = _request("GET", f"{_base_url()}/api/v1/platform/zone-paths",
                        params={"type": path_type}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Orders â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_available_orders(path: str = None):
    params = {"path": path} if path else {}
    resp = _request("GET", f"{_base_url()}/api/v1/demo/orders/available",
                        params=params, headers=_headers())
    return _handle_response(resp)


def get_assigned_orders(path: str = None):
    params = {"path": path} if path else {}
    resp = _request("GET", f"{_base_url()}/api/v1/worker/orders/assigned",
                        params=params, headers=_headers())
    return _handle_response(resp)


def get_all_orders(path: str = None):
    params = {"path": path} if path else {}
    resp = _request("GET", f"{_base_url()}/api/v1/worker/orders",
                        params=params, headers=_headers())
    return _handle_response(resp)


def get_order_detail(order_id: str):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/orders/{order_id}",
                        headers=_headers())
    return _handle_response(resp)


def accept_order(order_id: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/orders/{order_id}/accept",
                        headers=_headers())
    return _handle_response(resp)


def picked_up_order(order_id: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/orders/{order_id}/picked-up",
                        headers=_headers())
    return _handle_response(resp)


def delivered_order(order_id: str, customer_code: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/orders/{order_id}/delivered",
                        params={"customer_code": customer_code}, headers=_headers())
    return _handle_response(resp)


def send_customer_code(order_id: str):
    resp = _request("POST", f"{_base_url()}/api/v1/worker/orders/{order_id}/code/send",
                         headers=_headers())
    return _handle_response(resp)


# â”€â”€ Batches â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_available_batches(limit: int = 100):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/batches",
                        params={"limit": limit}, headers=_headers())
    return _handle_response(resp)


def get_assigned_batches():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/batches/assigned",
                        headers=_headers())
    return _handle_response(resp)


def get_delivered_batches():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/batches/delivered",
                        headers=_headers())
    return _handle_response(resp)


def accept_batch(batch_id: str, order_ids: list, pickup_code: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/batches/{batch_id}/accept",
                         json={"order_ids": order_ids, "pickup_code": pickup_code},
                         headers=_headers())
    return _handle_response(resp)


def deliver_batch(batch_id: str, delivery_code: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/batches/{batch_id}/deliver",
                        json={"delivery_code": delivery_code}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Fetch Verification â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def send_verification_code():
    resp = _request("POST", f"{_base_url()}/api/v1/worker/fetch-verification/send-code",
                         headers=_headers())
    return _handle_response(resp)


def verify_code(code: str):
    resp = _request("POST", f"{_base_url()}/api/v1/worker/fetch-verification/verify",
                         json={"code": code}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Session Tracking â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_session(session_id: str):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/session/{session_id}",
                        headers=_headers())
    return _handle_response(resp)


def get_session_deliveries(session_id: str):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/session/{session_id}/deliveries",
                        headers=_headers())
    return _handle_response(resp)


def get_session_fraud_signals(session_id: str):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/session/{session_id}/fraud-signals",
                        headers=_headers())
    return _handle_response(resp)


def end_session(session_id: str):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/session/{session_id}/end",
                        headers=_headers())
    return _handle_response(resp)


# â”€â”€ Earnings â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_earnings():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/earnings", headers=_headers())
    return _handle_response(resp)


def get_earnings_history():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/earnings/history", headers=_headers())
    return _handle_response(resp)


def get_baseline():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/earnings/baseline", headers=_headers())
    return _handle_response(resp)


# â”€â”€ Policy â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_policy():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/policy", headers=_headers())
    return _handle_response(resp)


def get_premium():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/policy/premium", headers=_headers())
    return _handle_response(resp)


def enroll_policy():
    resp = _request("POST", f"{_base_url()}/api/v1/worker/policy/enroll", headers=_headers())
    return _handle_response(resp)


def pay_premium(amount: int = None):
    payload = {}
    if amount is not None:
        payload["amount"] = amount
    resp = _request("POST", f"{_base_url()}/api/v1/worker/policy/premium/pay",
                         json=payload, headers=_headers())
    return _handle_response(resp)


def pause_policy():
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/policy/pause", headers=_headers())
    return _handle_response(resp)


def cancel_policy():
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/policy/cancel", headers=_headers())
    return _handle_response(resp)


# â”€â”€ Plans â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_plans():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/plans", headers=_headers())
    return _handle_response(resp)


def select_plan(plan_id: str, payment_amount_inr: int, expected_deliveries: int = None):
    payload = {"plan_id": plan_id, "payment_amount_inr": payment_amount_inr,
               "payment_confirmed": True}
    if expected_deliveries is not None:
        payload["expected_deliveries"] = expected_deliveries
    resp = _request("POST", f"{_base_url()}/api/v1/worker/plans/select",
                         json=payload, headers=_headers())
    return _handle_response(resp)


def skip_plan():
    resp = _request("POST", f"{_base_url()}/api/v1/worker/plans/skip", headers=_headers())
    return _handle_response(resp)


# â”€â”€ Claims & Wallet â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_claims():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/claims", headers=_headers())
    return _handle_response(resp)


def get_claim_detail(claim_id: str):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/claims/{claim_id}", headers=_headers())
    return _handle_response(resp)


def get_wallet():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/wallet", headers=_headers())
    return _handle_response(resp)


def get_payouts(limit: int = 10):
    resp = _request("GET", f"{_base_url()}/api/v1/worker/payouts",
                        params={"limit": limit}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Notifications â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def get_notifications():
    resp = _request("GET", f"{_base_url()}/api/v1/worker/notifications", headers=_headers())
    return _handle_response(resp)


def update_notification_preferences(prefs: dict):
    resp = _request("PUT", f"{_base_url()}/api/v1/worker/notifications/preferences",
                        json=prefs, headers=_headers())
    return _handle_response(resp)


def update_fcm_token(fcm_token: str):
    resp = _request("POST", f"{_base_url()}/api/v1/worker/notifications/fcm-token",
                         json={"fcm_token": fcm_token}, headers=_headers())
    return _handle_response(resp)


# â”€â”€ Demo / Dev Tools â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

def trigger_disruption(disruption_type: str, zone_level: str, zone_name: str):
    resp = _request("POST", f"{_base_url()}/api/v1/demo/trigger-disruption",
                         json={"disruption_type": disruption_type,
                               "zone_level": zone_level, "zone_name": zone_name},
                         headers=_headers())
    return _handle_response(resp)


def assign_orders(count: int):
    resp = _request("POST", f"{_base_url()}/api/v1/demo/assign-orders",
                         json={"count": count}, headers=_headers())
    return _handle_response(resp)


def simulate_deliveries(count: int):
    resp = _request("POST", f"{_base_url()}/api/v1/demo/simulate-deliveries",
                         json={"count": count}, headers=_headers())
    return _handle_response(resp)


def settle_earnings():
    resp = _request("POST", f"{_base_url()}/api/v1/demo/settle-earnings", headers=_headers())
    return _handle_response(resp)


def reset_zone():
    resp = _request("POST", f"{_base_url()}/api/v1/demo/reset-zone", headers=_headers())
    return _handle_response(resp)


def reset_demo():
    resp = _request("POST", f"{_base_url()}/api/v1/demo/reset", headers=_headers())
    return _handle_response(resp)

