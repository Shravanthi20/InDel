"""Platform blueprint for zone lookups and diagnostics."""
from flask import Blueprint, jsonify

from app.services import api_client
from app.services.api_client import ApiError
from app.session.store import login_required

platform_bp = Blueprint("platform", __name__)


@platform_bp.route("/platform/zones", methods=["GET"])
@login_required
def zones():
    """Return platform zones as JSON (used by onboarding/profile flows)."""
    try:
        return jsonify(api_client.get_zones()), 200
    except ApiError as e:
        return jsonify({"error": e.message}), e.status_code
