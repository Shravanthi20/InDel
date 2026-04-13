"""Auth blueprint — Login, Register, OTP, Onboarding, Plan Selection routes."""
from flask import Blueprint, render_template, request, redirect, url_for, flash, session as flask_session, jsonify
from app.services import api_client
from app.services.api_client import ApiError
from app.session.store import save_auth, login_required
from app.constants import ZONE_LEVELS, get_zone_names_for_level, ALL_VEHICLES, get_vehicles_for_level, is_valid_upi

auth_bp = Blueprint("auth", __name__)


@auth_bp.route("/login", methods=["GET", "POST"])
def login():
    if flask_session.get("token"):
        return redirect(url_for("worker.home"))
    error = None
    if request.method == "POST":
        identifier = request.form.get("identifier", "").strip()
        password = request.form.get("password", "").strip()
        email = identifier if "@" in identifier else None
        phone = identifier if "@" not in identifier else None
        try:
            data = api_client.login(password=password, email=email, phone=phone)
            save_auth(data["token"], data["worker_id"], data.get("token_type", "Bearer"))
            return redirect(url_for("worker.landing"))
        except ApiError as e:
            error = e.message
    return render_template("auth/login.html", error=error)


@auth_bp.route("/register", methods=["GET", "POST"])
def register():
    error = None
    form_data = {
        "username": "",
        "phone": "",
        "email": "",
        "password": "",
        "confirm_password": "",
        "zone_level": "",
        "zone_name": "",
    }
    if request.method == "POST":
        form_data = {
            "username": request.form.get("username", "").strip(),
            "phone": request.form.get("phone", "").strip(),
            "email": request.form.get("email", "").strip(),
            "password": request.form.get("password", "").strip(),
            "confirm_password": request.form.get("confirm_password", "").strip(),
            "zone_level": request.form.get("zone_level", "").strip().upper(),
            "zone_name": request.form.get("zone_name", "").strip(),
        }
        zone_level = form_data["zone_level"]
        zone_name = form_data["zone_name"]
        try:
            if form_data["password"] != form_data["confirm_password"]:
                raise ApiError(400, "Passwords do not match")
            if zone_level not in {"A", "B", "C"}:
                raise ApiError(400, "Zone level is required.")
            zone_paths = api_client.get_zone_paths(zone_level.lower())
            available_names = []
            if zone_level == "A":
                available_names = zone_paths.get("cities", [])
            else:
                for pair in zone_paths.get("city_pairs", []):
                    if zone_level == "B":
                        available_names.append(f"{pair.get('from')} to {pair.get('to')} ({pair.get('state') or ''})".strip())
                    else:
                        available_names.append(
                            f"{pair.get('from')} ({pair.get('from_state') or ''}) to {pair.get('to')} ({pair.get('to_state') or ''})".strip()
                        )
            if zone_name not in available_names:
                raise ApiError(400, "Please select a valid zone name.")
            data = api_client.register(
                username=form_data["username"],
                phone=form_data["phone"],
                email=form_data["email"],
                password=form_data["password"],
                zone_level=zone_level,
                zone_name=zone_name,
            )
            save_auth(data["token"], data["worker_id"])
            return redirect(url_for("auth.onboarding"))
        except ApiError as e:
            error = e.message
    return render_template("auth/register.html", error=error, form_data=form_data, zone_levels=ZONE_LEVELS)


@auth_bp.route("/otp", methods=["GET", "POST"])
def otp():
    """OTP recovery / fallback login."""
    error = None
    otp_sent = False
    phone = request.args.get("phone", "")
    if request.method == "POST":
        action = request.form.get("action")
        phone = request.form.get("phone", "").strip()
        if action == "send":
            try:
                api_client.send_otp(phone)
                otp_sent = True
            except ApiError as e:
                error = e.message
        elif action == "verify":
            otp_code = request.form.get("otp", "").strip()
            try:
                data = api_client.verify_otp(phone, otp_code)
                save_auth(data["token"], data["worker_id"])
                return redirect(url_for("worker.landing"))
            except ApiError as e:
                error = e.message
    return render_template("auth/otp.html", error=error, otp_sent=otp_sent, phone=phone)


@auth_bp.route("/onboarding", methods=["GET", "POST"])
@login_required
def onboarding():
    error = None
    form_data = {
        "name": "",
        "vehicle_type": "",
        "vehicle_name": "",
        "upi_id": "",
    }

    if request.method == "POST":
        form_data = {
            "name": request.form.get("name", "").strip(),
            "vehicle_type": request.form.get("vehicle_type", "").strip(),
            "vehicle_name": request.form.get("vehicle_name", "").strip(),
            "upi_id": request.form.get("upi_id", "").strip(),
        }
        if not form_data["name"]:
            error = "Please fill all fields"
        elif not is_valid_upi(form_data["upi_id"]):
            error = "Invalid UPI ID format (username@bankid)"
        elif not form_data["vehicle_type"] or not form_data["vehicle_name"]:
            error = "Please fill all fields"
        if not error:
            try:
                api_client.onboard(
                    name=form_data["name"],
                    vehicle_type=form_data["vehicle_type"],
                    vehicle_name=form_data["vehicle_name"],
                    upi_id=form_data["upi_id"],
                )
                return redirect(url_for("auth.plan_selection"))
            except ApiError as e:
                error = e.message

    vehicle_type_options = ["two-wheeler", "four-wheeler-small", "four-wheeler-large"]
    common_transport_means = [
        "scooter",
        "motorcycle",
        "auto-rickshaw",
        "hatchback",
        "sedan",
        "suv",
        "pickup-van",
        "mini-truck",
        "other",
    ]

    return render_template(
        "auth/onboarding.html",
        error=error,
        form_data=form_data,
        vehicle_type_options=vehicle_type_options,
        common_transport_means=common_transport_means,
    )


@auth_bp.route("/plan-selection", methods=["GET", "POST"])
@login_required
def plan_selection():
    error = None
    plans = []
    try:
        plans_data = api_client.get_plans()
        plans = plans_data.get("plans", [])
    except ApiError as e:
        error = e.message

    if request.method == "POST":
        action = request.form.get("action")
        if action == "skip":
            try:
                api_client.skip_plan()
                return redirect(url_for("worker.landing"))
            except ApiError as e:
                error = e.message
        else:
            try:
                api_client.select_plan(
                    plan_id=request.form.get("plan_id", ""),
                    payment_amount_inr=int(request.form.get("payment_amount_inr", 0)),
                    expected_deliveries=int(d) if (d := request.form.get("expected_deliveries")) else None,
                )
                return redirect(url_for("worker.landing"))
            except ApiError as e:
                error = e.message

    return render_template("plan/plan_selection.html", error=error, plans=plans)


@auth_bp.route("/logout")
def logout():
    from app.session.store import clear_auth
    clear_auth()
    return redirect(url_for("auth.login"))


@auth_bp.route("/zone-paths", methods=["GET"])
def zone_paths():
    path_type = (request.args.get("type") or "").strip().lower()
    if path_type not in {"a", "b", "c"}:
        return jsonify({"error": "invalid_type"}), 400
    try:
        return jsonify(api_client.get_zone_paths(path_type))
    except ApiError as e:
        return jsonify({"error": e.message}), e.status_code
