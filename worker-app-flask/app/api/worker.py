"""
Worker blueprint — all main app screens.
Mirrors the Kotlin NavGraph: Landing, Home, Orders, Delivery flow,
Earnings, Policy, Claims, Payouts, Notifications, Profile.
"""
from flask import Blueprint, render_template, request, redirect, url_for, session as flask_session
from app.services import api_client
from app.services.api_client import ApiError
from app.session.store import login_required
from app.constants import ZONE_LEVELS, get_zone_names_for_level, ALL_VEHICLES, get_vehicles_for_level

worker_bp = Blueprint("worker", __name__)


# ── Landing ───────────────────────────────────────────────────────────────────

@worker_bp.route("/landing")
@login_required
def landing():
    profile, earnings, policy = {}, {}, {}
    error = None
    try:
        profile = api_client.get_profile().get("worker", {})
        earnings = api_client.get_earnings()
        policy = api_client.get_policy().get("policy", {})
    except ApiError as e:
        if e.status_code == 401:
            return redirect(url_for("auth.login"))
        error = e.message
    return render_template("delivery/landing.html",
                           profile=profile, earnings=earnings,
                           policy=policy, error=error)


# ── Home Dashboard ────────────────────────────────────────────────────────────

@worker_bp.route("/home")
@login_required
def home():
    profile, earnings, policy = {}, {}, {}
    error = None
    try:
        profile = api_client.get_profile().get("worker", {})
        earnings = api_client.get_earnings()
        policy = api_client.get_policy().get("policy", {})
    except ApiError as e:
        if e.status_code == 401:
            return redirect(url_for("auth.login"))
        error = e.message
    has_disruption = policy.get("status") == "disrupted"
    return render_template("home/dashboard.html",
                           profile=profile, earnings=earnings,
                           policy=policy, has_disruption=has_disruption, error=error)


# ── Orders ────────────────────────────────────────────────────────────────────

@worker_bp.route("/orders")
@login_required
def orders():
    available_orders, assigned_orders = [], []
    error = None
    try:
        available_data = api_client.get_available_orders()
        available_orders = available_data.get("orders", [])
    except ApiError as e:
        error = e.message
    try:
        assigned_data = api_client.get_assigned_orders()
        assigned_orders = assigned_data.get("orders", [])
    except ApiError:
        pass
    return render_template("delivery/orders.html",
                           available_orders=available_orders,
                           assigned_orders=assigned_orders,
                           error=error)


@worker_bp.route("/orders/<order_id>/accept", methods=["POST"])
@login_required
def accept_order(order_id):
    try:
        api_client.accept_order(order_id)
    except ApiError:
        pass
    return redirect(url_for("worker.fetch_verification") + f"?order_id={order_id}")


# ── Batch Detail ──────────────────────────────────────────────────────────────

@worker_bp.route("/batch/<batch_id>")
@login_required
def batch_detail(batch_id):
    batches, batch = [], None
    try:
        data = api_client.get_assigned_batches()
        batches = data.get("batches", [])
        batch = next((b for b in batches if b.get("batch_id") == batch_id), None)
    except ApiError as e:
        return render_template("delivery/orders.html", error=e.message)
    return render_template("delivery/batch_detail.html", batch=batch, batch_id=batch_id)


# ── Fetch Verification ────────────────────────────────────────────────────────

@worker_bp.route("/fetch-verification", methods=["GET", "POST"])
@login_required
def fetch_verification():
    error = None
    success = None
    order_id = request.args.get("order_id", "")
    zone_config = {}
    try:
        zone_config = api_client.get_zone_config()
    except ApiError:
        pass

    if request.method == "POST":
        action = request.form.get("action")
        order_id = request.form.get("order_id", "")
        if action == "send_code":
            try:
                api_client.send_verification_code()
                success = "Verification code sent to your phone."
            except ApiError as e:
                error = e.message
        elif action == "verify":
            code = request.form.get("code", "").strip()
            try:
                api_client.verify_code(code)
                return redirect(url_for("worker.delivery_execution", order_id=order_id))
            except ApiError as e:
                error = e.message

    return render_template("delivery/fetch_verification.html",
                           error=error, success=success,
                           order_id=order_id, zone_config=zone_config)


# ── Delivery Execution ────────────────────────────────────────────────────────

@worker_bp.route("/delivery/<order_id>", methods=["GET", "POST"])
@login_required
def delivery_execution(order_id):
    order = {}
    error = None
    try:
        order = api_client.get_order_detail(order_id)
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        try:
            api_client.picked_up_order(order_id)
            api_client.send_customer_code(order_id)
            return redirect(url_for("worker.delivery_completion", order_id=order_id))
        except ApiError as e:
            error = e.message

    return render_template("delivery/execution.html", order=order,
                           order_id=order_id, error=error)


# ── Delivery Completion ───────────────────────────────────────────────────────

@worker_bp.route("/delivery/<order_id>/complete", methods=["GET", "POST"])
@login_required
def delivery_completion(order_id):
    order = {}
    error = None
    success = None
    try:
        order = api_client.get_order_detail(order_id)
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        customer_code = request.form.get("customer_code", "").strip()
        try:
            api_client.delivered_order(order_id, customer_code)
            success = f"Delivery confirmed! ₹{order.get('earning_inr', 0):.0f} earned."
        except ApiError as e:
            error = e.message

    return render_template("delivery/completion.html", order=order,
                           order_id=order_id, error=error, success=success)


# ── Session Tracking ──────────────────────────────────────────────────────────

@worker_bp.route("/session", methods=["GET", "POST"])
@login_required
def session_tracking():
    sess, deliveries, fraud_signals = {}, [], []
    error = None
    # Use session_id stored in Flask session if available
    session_id = flask_session.get("delivery_session_id", "current")

    if request.method == "POST" and request.form.get("action") == "end":
        try:
            api_client.end_session(session_id)
            flask_session.pop("delivery_session_id", None)
        except ApiError as e:
            error = e.message

    try:
        sess = api_client.get_session(session_id)
        deliveries = api_client.get_session_deliveries(session_id).get("orders", [])
        fraud_signals = api_client.get_session_fraud_signals(session_id).get("signals", [])
    except ApiError as e:
        error = e.message

    return render_template("delivery/session_tracking.html",
                           sess=sess, deliveries=deliveries,
                           fraud_signals=fraud_signals, error=error)


# ── Earnings ──────────────────────────────────────────────────────────────────

@worker_bp.route("/earnings")
@login_required
def earnings():
    earnings_data, history, baseline = {}, [], {}
    error = None
    try:
        earnings_data = api_client.get_earnings()
        history = api_client.get_earnings_history().get("history", [])
        baseline = api_client.get_baseline()
    except ApiError as e:
        error = e.message
    return render_template("earnings/earnings.html",
                           earnings=earnings_data, history=history,
                           baseline=baseline, error=error)


# ── Policy ────────────────────────────────────────────────────────────────────

@worker_bp.route("/policy", methods=["GET", "POST"])
@login_required
def policy():
    policy_data, premium = {}, {}
    error = None
    success = None

    try:
        policy_data = api_client.get_policy().get("policy", {})
        premium = api_client.get_premium()
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        action = request.form.get("action")
        try:
            if action == "enroll":
                api_client.enroll_policy()
                success = "Successfully enrolled in policy."
            elif action == "pause":
                api_client.pause_policy()
                success = "Policy paused."
            elif action == "cancel":
                api_client.cancel_policy()
                success = "Policy cancelled."
            policy_data = api_client.get_policy().get("policy", {})
        except ApiError as e:
            error = e.message

    return render_template("policy/policy.html",
                           policy=policy_data, premium=premium,
                           error=error, success=success)


@worker_bp.route("/policy/premium-pay", methods=["GET", "POST"])
@login_required
def premium_pay():
    premium = {}
    error = None
    success = None
    try:
        premium = api_client.get_premium()
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        amount = request.form.get("amount")
        try:
            api_client.pay_premium(int(amount) if amount else None)
            success = "Premium payment successful!"
        except ApiError as e:
            error = e.message

    return render_template("policy/premium_pay.html",
                           premium=premium, error=error, success=success)


# ── Claims ────────────────────────────────────────────────────────────────────

@worker_bp.route("/claims")
@login_required
def claims():
    claims_list = []
    error = None
    try:
        data = api_client.get_claims()
        claims_list = data.get("claims", [])
    except ApiError as e:
        error = e.message
    return render_template("claims/claims.html", claims=claims_list, error=error)


@worker_bp.route("/claims/<claim_id>")
@login_required
def claim_detail(claim_id):
    claim = {}
    error = None
    try:
        claim = api_client.get_claim_detail(claim_id)
    except ApiError as e:
        error = e.message
    return render_template("claims/claim_detail.html", claim=claim, error=error)


# ── Payouts ───────────────────────────────────────────────────────────────────

@worker_bp.route("/payouts")
@login_required
def payouts():
    wallet, payouts_list = {}, []
    error = None
    try:
        wallet = api_client.get_wallet()
        payouts_list = api_client.get_payouts().get("payouts", [])
    except ApiError as e:
        error = e.message
    return render_template("payouts/payouts.html",
                           wallet=wallet, payouts=payouts_list, error=error)


# ── Notifications ─────────────────────────────────────────────────────────────

@worker_bp.route("/notifications", methods=["GET", "POST"])
@login_required
def notifications():
    notifs = []
    error = None
    success = None
    try:
        notifs = api_client.get_notifications().get("notifications", [])
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        prefs = {k: (v == "on") for k, v in request.form.items() if k != "action"}
        try:
            api_client.update_notification_preferences(prefs)
            success = "Preferences saved."
        except ApiError as e:
            error = e.message

    return render_template("notifications/notifications.html",
                           notifications=notifs, error=error, success=success)


# ── Profile Edit ──────────────────────────────────────────────────────────────

@worker_bp.route("/profile", methods=["GET", "POST"])
@login_required
def profile_edit():
    profile = {}
    error = None
    success = None
    zone_names_by_level = {
        level["level"]: get_zone_names_for_level(level["level"]) for level in ZONE_LEVELS
    }
    vehicles_by_level = {level["level"]: get_vehicles_for_level(level["level"]) for level in ZONE_LEVELS}
    try:
        profile = api_client.get_profile().get("worker", {})
    except ApiError as e:
        error = e.message

    form_data = {
        "name": profile.get("name", ""),
        "zone_level": (profile.get("zone_level") or "").strip().upper(),
        "zone_name": profile.get("zone_name", ""),
        "city": profile.get("city", ""),
        "vehicle_type": profile.get("vehicle_type", ""),
        "upi_id": profile.get("upi_id", ""),
    }

    if request.method == "POST":
        form_data = {
            "name": request.form.get("name", "").strip(),
            "zone_level": request.form.get("zone_level", "").strip().upper(),
            "zone_name": request.form.get("zone_name", "").strip(),
            "city": request.form.get("city", "").strip(),
            "vehicle_type": request.form.get("vehicle_type", "").strip(),
            "upi_id": request.form.get("upi_id", "").strip(),
        }

        zone_level = form_data["zone_level"]
        zone_name = form_data["zone_name"]
        vehicle_type = form_data["vehicle_type"]
        
        valid_levels = [opt["level"] for opt in ZONE_LEVELS]
        if not zone_level:
            error = "Zone level is required."
        elif zone_level not in valid_levels:
            error = f"Invalid zone level. Choose from: {', '.join(valid_levels)}"
        else:
            try:
                zone_paths = api_client.get_zone_paths(zone_level.lower())
                valid_names = []
                if zone_level == "A":
                    valid_names = zone_paths.get("cities", [])
                else:
                    for pair in zone_paths.get("city_pairs", []):
                        if zone_level == "B":
                            valid_names.append(
                                f"{pair.get('from')} to {pair.get('to')} ({pair.get('state') or ''})".strip()
                            )
                        else:
                            valid_names.append(
                                f"{pair.get('from')} ({pair.get('from_state') or ''}) to {pair.get('to')} ({pair.get('to_state') or ''})".strip()
                            )
                if zone_name not in valid_names:
                    error = f"Invalid zone name for level {zone_level}. Choose from: {', '.join(valid_names)}"
                allowed_vehicles = get_vehicles_for_level(zone_level)
                if vehicle_type not in allowed_vehicles:
                    error = f"Invalid vehicle for level {zone_level}. Choose from: {', '.join(allowed_vehicles)}"
            except ApiError as e:
                error = e.message
        
        if not error:
            try:
                api_client.update_profile(
                    name=form_data["name"],
                    zone_level=zone_level or None,
                    zone_name=zone_name or None,
                    vehicle_type=form_data["vehicle_type"],
                    upi_id=form_data["upi_id"],
                    city=form_data["city"] or None,
                    from_city=request.form.get("from_city") or None,
                    to_city=request.form.get("to_city") or None,
                )
                profile = api_client.get_profile().get("worker", {})
                form_data = {
                    "name": profile.get("name", ""),
                    "zone_level": (profile.get("zone_level") or "").strip().upper(),
                    "zone_name": profile.get("zone_name", ""),
                    "city": profile.get("city", ""),
                    "vehicle_type": profile.get("vehicle_type", ""),
                    "upi_id": profile.get("upi_id", ""),
                }
                success = "Profile updated successfully."
            except ApiError as e:
                error = e.message

    selected_level = form_data["zone_level"]
    selected_zone_names = get_zone_names_for_level(selected_level) if selected_level else []
    vehicles = get_vehicles_for_level(selected_level) if selected_level else ALL_VEHICLES

    return render_template("profile/profile_edit.html",
                           profile=profile, error=error, success=success,
                           zone_levels=ZONE_LEVELS, vehicles=vehicles,
                           form_data=form_data,
                           selected_zone_names=selected_zone_names,
                           zone_names_by_level=zone_names_by_level,
                           vehicles_by_level=vehicles_by_level)
