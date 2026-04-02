from fastapi import FastAPI, HTTPException
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
    from ml.premium.model import PremiumModel
    from ml.premium.shap_explainer import SHAPExplainer

app = FastAPI(title="InDel Premium Prediction Service - V1")

# --- Schemas ---

class PremiumRequest(BaseModel):
    worker_id: str
    zone_id: str
    vehicle_type: str
    season: str
    recent_disruption_rate: float
    order_volatility: float
    rainfall_mm: float
    temp_c: float
    aqi: float

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
MODEL_PATH = 'c:/Users/ritha/OneDrive/Documents/Amrita/Devtrails/InDel/ml/premium/artifacts/premium_model.joblib'
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

@app.on_event("startup")
def startup_event():
    load_model_instance()

# --- Endpoints ---

@app.get("/health")
def health():
    return {"status": "ok", "service": "premium-ml", "model_loaded": model is not None}

@app.post("/ml/v1/premium/calculate", response_model=PremiumResponse)
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
    
    response_data = PremiumData(
        worker_id=request.worker_id,
        premium_inr=round(float(premium[0]), 2),
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

@app.post("/ml/v1/premium/batch-calculate", response_model=BatchPremiumResponse)
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
        results.append(PremiumData(
            worker_id=request.worker_id,
            premium_inr=round(float(premiums[i]), 2),
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
