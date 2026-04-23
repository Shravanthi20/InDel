from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional
import pandas as pd
import uuid
from datetime import datetime
import os
import sys

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
    
    # Enforce actuarial floor/ceiling (e.g., Rs. 49 base minimum for new out-of-distribution workers)
    final_premium = round(float(premium[0]), 2)
    final_premium = max(49.0, min(250.0, final_premium))
    
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
        final_premium = round(float(premiums[i]), 2)
        final_premium = max(49.0, min(250.0, final_premium))
        
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
    uvicorn.run(app, host="0.0.0.0", port=8000)
