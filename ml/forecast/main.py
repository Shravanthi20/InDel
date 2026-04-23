import os
import logging
from contextlib import asynccontextmanager
from datetime import date, timedelta

import pandas as pd
from fastapi import APIRouter, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

try:
    from prophet_model import ProphetForecaster
except ImportError:
    from forecast.prophet_model import ProphetForecaster

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("forecast-ml")

# ── Configuration ────────────────────────────────────────────────────────────

SUPPORTED_ZONES = [1, 2, 3, 4]

DATA_PATH = os.path.join(os.path.dirname(__file__), "data", "synthetic_zone_history.csv")

MODEL_INFO = {
    "model": "Facebook Prophet (Weekly Seasonal, pandas backend)",
    "scope": "Per-zone time-series. No cross-zone correlation modelled.",
    "purpose": "Reserve planning only. Does not influence claim approval or individual claim decisioning.",
    "retraining_cadence": "Weekly — every Monday 02:00 UTC",
    "selected_because": "Reliable on small datasets; strong weekly seasonality decomposition; interpretable uncertainty intervals.",
    "known_limitations": [
        "Per-zone only — cannot model correlated multi-zone catastrophes.",
        "Uncertainty intervals widen significantly beyond 3-day horizon.",
        "Prophet Stan backend replaced with pandas seasonal model for container compatibility.",
        "Trained on synthetic zone history (real data accumulation in progress).",
    ],
    "upgrade_path": "DeepAR (AWS) for cross-zone joint distribution modelling.",
    "mitigation": "Reinsurance layer and catastrophic cap at portfolio level.",
}

# Fallback static profiles — used if CSV is missing or Prophet fails
FALLBACK_PROFILES = {
    1: [0.28, 0.31, 0.19, 0.41, 0.22, 0.35, 0.18],
    2: [0.14, 0.17, 0.12, 0.21, 0.09, 0.16, 0.11],
    3: [0.22, 0.26, 0.20, 0.33, 0.18, 0.29, 0.15],
    4: [0.09, 0.12, 0.08, 0.15, 0.07, 0.11, 0.08],
}

# ── Model cache (trained at startup) ─────────────────────────────────────────

_forecasters: dict[int, ProphetForecaster] = {}
_using_fallback: dict[int, bool] = {}


def train_all_zones():
    """Load CSV and train a Prophet model per zone at startup."""
    if not os.path.exists(DATA_PATH):
        log.warning(f"Training data not found at {DATA_PATH}. All zones will use fallback profiles.")
        for z in SUPPORTED_ZONES:
            _using_fallback[z] = True
        return

    df = pd.read_csv(DATA_PATH)
    log.info(f"Loaded {len(df)} rows from {DATA_PATH}")

    for zone_id in SUPPORTED_ZONES:
        zone_data = df[df["zone_id"] == zone_id].copy()
        log.info(f"Zone {zone_id}: {len(zone_data)} training rows")

        forecaster = ProphetForecaster(zone_id)
        try:
            if len(zone_data) < 10:
                raise ValueError(f"Insufficient data ({len(zone_data)} rows); need ≥10.")
            forecaster.train(zone_data)
            _forecasters[zone_id] = forecaster
            _using_fallback[zone_id] = False
            log.info(f"Zone {zone_id}: Prophet model trained ✅")
        except Exception as e:
            log.warning(f"Zone {zone_id}: Prophet training failed — using fallback. Reason: {e}")
            _using_fallback[zone_id] = True


# ── FastAPI app ───────────────────────────────────────────────────────────────

router = APIRouter()


# ── Schemas ───────────────────────────────────────────────────────────────────

class ForecastRequest(BaseModel):
    zone_id: int


class ForecastPoint(BaseModel):
    date: str
    disruption_probability: float


class ForecastResponse(BaseModel):
    zone_id: int
    model: str
    purpose: str
    retraining_cadence: str
    scope: str
    inference: str          # "prophet" or "fallback"
    forecast: list[ForecastPoint]


# ── Routes ─────────────────────────────────────────────────────────────────────

@router.get("/health")
@router.get("/forecast/health")
def health_forecast():
    return {
        "status": "ok",
        "service": "forecast-ml",
        "model": "prophet",
        "zones_trained": [z for z in SUPPORTED_ZONES if not _using_fallback.get(z, True)],
        "zones_fallback": [z for z in SUPPORTED_ZONES if _using_fallback.get(z, True)],
    }


@router.get("/forecast/model-info")
def model_info():
    """Explicit model metadata: scope, limitations, retraining cadence, upgrade path."""
    return MODEL_INFO


@router.get("/forecast/zones")
def available_zones():
    """Lists zones with per-zone model status."""
    return {
        "zones": SUPPORTED_ZONES,
        "note": "Each zone has an independent Prophet model. Cross-zone correlation is not modelled.",
        "status": {
            z: ("prophet" if not _using_fallback.get(z, True) else "fallback")
            for z in SUPPORTED_ZONES
        },
    }


@router.post("/forecast", response_model=ForecastResponse)
def generate_forecast(request: ForecastRequest):
    """
    Returns a 7-day disruption probability forecast for the given zone.

    IMPORTANT: Reserve planning only.
    This output does NOT determine whether an individual claim is approved.
    Prophet is trained per-zone on historical disruption event frequency.
    Retraining cadence: weekly (every Monday 02:00 UTC).
    """
    if request.zone_id not in SUPPORTED_ZONES:
        raise HTTPException(
            status_code=404,
            detail=(
                f"No forecast model available for zone_id={request.zone_id}. "
                f"Supported zones: {SUPPORTED_ZONES}"
            ),
        )

    using_fallback = _using_fallback.get(request.zone_id, True)

    if not using_fallback:
        # Real Prophet inference
        forecaster = _forecasters[request.zone_id]
        raw = forecaster.forecast(periods=7)
        forecast_points = [ForecastPoint(**p) for p in raw]
        inference = "seasonal"
    else:
        # Fallback: static profiles with dynamic dates
        profile = FALLBACK_PROFILES[request.zone_id]
        today = date.today()
        forecast_points = [
            ForecastPoint(
                date=(today + timedelta(days=i)).isoformat(),
                disruption_probability=round(profile[i], 3),
            )
            for i in range(7)
        ]
        inference = "fallback"

    log.info(
        f"Forecast zone={request.zone_id} inference={inference} "
        f"probs={[p.disruption_probability for p in forecast_points]}"
    )

    return ForecastResponse(
        zone_id=request.zone_id,
        model=MODEL_INFO["model"],
        purpose=MODEL_INFO["purpose"],
        retraining_cadence=MODEL_INFO["retraining_cadence"],
        scope=MODEL_INFO["scope"],
        inference=inference,
        forecast=forecast_points,
    )


if __name__ == "__main__":
    from fastapi import FastAPI
    import uvicorn
    app = FastAPI()
    app.include_router(router)
    uvicorn.run(app, host="0.0.0.0", port=8000)
