from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional
import pandas as pd
import uuid
from datetime import datetime
import os
import sys
import math

# Ensure local imports work
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
try:
    from model import PremiumModel
    from shap_explainer import SHAPExplainer
except ImportError:
    # Handle the case where the script is run from a different directory
    from premium.model import PremiumModel
    from premium.shap_explainer import SHAPExplainer

from fastapi.middleware.cors import CORSMiddleware

router = APIRouter()

# --- Schemas ---

class PremiumRequest(BaseModel):
    worker_id: str
    zone_id: str
    city: str
    state: str
    zone_type: str
    vehicle_type: str
    season: str
    experience_days: int
    avg_daily_orders: float
    avg_daily_earnings: float
    active_hours_per_day: float
    rainfall_mm: float
    aqi: float
    temperature: float
    humidity: float
    order_volatility: float
    earnings_volatility: float
    recent_disruption_rate: float

class ExplainabilityFactor(BaseModel):
    feature: str
    impact: float

class PremiumData(BaseModel):
    worker_id: str
    premium_inr: float
    risk_score: float
    explainability: List[ExplainabilityFactor]
    model_version: str

class Meta(BaseModel):
    request_id: str
    timestamp: datetime

class PremiumResponse(BaseModel):
    data: PremiumData
    meta: Meta

class BatchPremiumResponse(BaseModel):
    data: List[PremiumData]
    meta: Meta

# --- Model Loading ---

model = None
explainer = None
script_dir = os.path.dirname(os.path.abspath(__file__))
MODEL_PATH = os.path.join(script_dir, 'artifacts/premium_model.joblib')
MODEL_VERSION = "premium_xgb_v1"


def _env_float(name: str, default: float) -> float:
    raw = os.getenv(name, "").strip()
    if not raw:
        return default
    try:
        return float(raw)
    except ValueError:
        return default


# Guardrails are configurable and intentionally broad so model predictions stay dynamic.
PREMIUM_MIN_INR = _env_float("PREMIUM_MIN_INR", 30.0)
PREMIUM_MAX_INR = _env_float("PREMIUM_MAX_INR", 120.0)
PREMIUM_BASE_INR = _env_float("PREMIUM_BASE_INR", PREMIUM_MIN_INR)
PREMIUM_RISK_BAND_INR = _env_float("PREMIUM_RISK_BAND_INR", 20.0)
PREMIUM_MODEL_WEIGHT = _env_float("PREMIUM_MODEL_WEIGHT", 0.30)
PREMIUM_REGION_STRESS_WEIGHT = _env_float("PREMIUM_REGION_STRESS_WEIGHT", 0.50)


def _clamp01(value: float) -> float:
    return max(0.0, min(1.0, value))


def apply_premium_guardrails(predicted: float) -> float:
    if not math.isfinite(predicted):
        return PREMIUM_MIN_INR

    lower = min(PREMIUM_MIN_INR, PREMIUM_MAX_INR)
    upper = max(PREMIUM_MIN_INR, PREMIUM_MAX_INR)
    return max(lower, min(upper, predicted))


def compute_dynamic_premium(predicted: float, risk_score: float, recent_disruption_rate: float) -> float:
    # Blend model output with risk-aware uplift so low-disruption regions remain cheaper
    # while high-disruption regions are priced higher, with a clear base floor.
    model_weight = _clamp01(PREMIUM_MODEL_WEIGHT)
    risk_component = PREMIUM_BASE_INR + (_clamp01(risk_score) * max(0.0, PREMIUM_RISK_BAND_INR))
    blended = (predicted * model_weight) + (risk_component * (1.0 - model_weight))

    # Region stress uplift: repeated disruptions in a region should increase premiums.
    # `recent_disruption_rate` is expected in [0, 1]. Example with weight=0.50:
    # rate=0.10 -> x1.05, rate=0.60 -> x1.30, rate=1.00 -> x1.50.
    region_stress = 1.0 + (_clamp01(recent_disruption_rate) * max(0.0, PREMIUM_REGION_STRESS_WEIGHT))
    blended = blended * region_stress

    return apply_premium_guardrails(blended)

def load_model_instance():
    global model, explainer
    if os.path.exists(MODEL_PATH):
        try:
            model = PremiumModel.load(MODEL_PATH)
            explainer = SHAPExplainer(model)
            print(f"Model {MODEL_VERSION} loaded successfully.")
        except Exception as e:
            print(f"Error loading model: {e}")
    else:
        print(f"Warning: Model artifacts not found at {MODEL_PATH}. Prediction endpoints will fail.")

# Startup handled in root

# --- Endpoints ---

@router.get("/health")
@router.get("/premium/health")
def health_premium():
    return {"status": "ok", "service": "premium-ml", "model_loaded": model is not None}

@router.post("/ml/v1/premium/calculate", response_model=PremiumResponse)
def calculate_premium(request: PremiumRequest):
    if model is None:
        # Try to reload
        load_model_instance()
        if model is None:
            raise HTTPException(status_code=503, detail="Model not loaded")
    
    # Convert request to DataFrame for model
    df = pd.DataFrame([request.dict()])
    
    # Predict
    premium, risk = model.predict(df)
    
    # Explain
    explainability = explainer.explain(df)[0]
    
    # Keep pricing dynamic; only apply broad guardrails to avoid invalid extremes.
    final_premium = round(
        compute_dynamic_premium(
            float(premium[0]),
            float(risk[0]),
            float(request.recent_disruption_rate),
        ),
        2,
    )
    
    response_data = PremiumData(
        worker_id=request.worker_id,
        premium_inr=final_premium,
        risk_score=round(float(risk[0]), 3),
        explainability=explainability,
        model_version=MODEL_VERSION
    )
    
    return PremiumResponse(
        data=response_data,
        meta=Meta(
            request_id=f"req_{uuid.uuid4().hex[:8]}",
            timestamp=datetime.utcnow()
        )
    )

@router.post("/ml/v1/premium/batch-calculate", response_model=BatchPremiumResponse)
def batch_calculate_premium(requests: List[PremiumRequest]):
    if model is None:
        load_model_instance()
        if model is None:
            raise HTTPException(status_code=503, detail="Model not loaded")
    
    # Convert all requests to DataFrame
    df = pd.DataFrame([r.dict() for r in requests])
    
    # Predict
    premiums, risks = model.predict(df)
    
    # Explain
    all_explainability = explainer.explain(df)
    
    results = []
    for i, request in enumerate(requests):
        final_premium = round(
            compute_dynamic_premium(
                float(premiums[i]),
                float(risks[i]),
                float(request.recent_disruption_rate),
            ),
            2,
        )
        
        results.append(PremiumData(
            worker_id=request.worker_id,
            premium_inr=final_premium,
            risk_score=round(float(risks[i]), 3),
            explainability=all_explainability[i],
            model_version=MODEL_VERSION
        ))
    
    return BatchPremiumResponse(
        data=results,
        meta=Meta(
            request_id=f"req_{uuid.uuid4().hex[:8]}",
            timestamp=datetime.utcnow()
        )
    )

if __name__ == "__main__":
    import uvicorn
    app_instance = FastAPI(title="InDel Premium ML Service")
    app_instance.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=False,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    app_instance.include_router(router)
    uvicorn.run(app_instance, host="0.0.0.0", port=8000)
