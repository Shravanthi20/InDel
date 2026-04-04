"""Demo / Dev Tools blueprint — mirrors Kotlin DevToolsScreen."""
from flask import Blueprint, render_template, request, redirect, url_for
from app.services import api_client
from app.services.api_client import ApiError
from app.session.store import login_required

demo_bp = Blueprint("demo", __name__)


@demo_bp.route("/dev-tools", methods=["GET", "POST"])
@login_required
def dev_tools():
    result = None
    error = None

    if request.method == "POST":
        action = request.form.get("action")
        try:
            if action == "trigger_disruption":
                api_client.trigger_disruption(
                    disruption_type=request.form.get("disruption_type", "WEATHER"),
                    zone_level=request.form.get("zone_level", "city"),
                    zone_name=request.form.get("zone_name", ""),
                )
                result = "Disruption triggered."
            elif action == "assign_orders":
                count = int(request.form.get("count", 3))
                api_client.assign_orders(count)
                result = f"{count} order(s) assigned."
            elif action == "simulate_deliveries":
                count = int(request.form.get("sim_count", 2))
                api_client.simulate_deliveries(count)
                result = f"{count} delivery(ies) simulated."
            elif action == "settle_earnings":
                api_client.settle_earnings()
                result = "Earnings settled."
            elif action == "reset_zone":
                api_client.reset_zone()
                result = "Zone reset."
            elif action == "reset_all":
                api_client.reset_demo()
                result = "Full reset complete."
        except ApiError as e:
            error = e.message

    return render_template("debug/dev_tools.html", result=result, error=error)
