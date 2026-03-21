from fastapi import FastAPI
from pydantic import BaseModel
import numpy as np

app = FastAPI(title="InDel Premium Prediction Service")

class PremiumRequest(BaseModel):
    worker_id: int
    earnings_volatility: float
    disruption_frequency: float
    zone_risk_rating: float
    account_age_days: int

class PremiumResponse(BaseModel):
    predicted_premium: float
    confidence: float
    explanation: str

@app.get("/health")
def health():
    return {"status": "ok", "service": "premium-ml"}

@app.post("/predict", response_model=PremiumResponse)
def predict_premium(request: PremiumRequest):
    # XGBoost prediction + SHAP explanation
    predicted_premium = 300.0  # Placeholder
    confidence = 0.92
    explanation = "Premium based on earnings stability and zone risk"
    
    return PremiumResponse(
        predicted_premium=predicted_premium,
        confidence=confidence,
        explanation=explanation
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
