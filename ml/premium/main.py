from fastapi import FastAPI
from pydantic import BaseModel

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


class PremiumV1Response(BaseModel):
    worker_id: int
    weekly_premium_inr: float
    risk_score: float
    confidence: float
    shap_breakdown: list[dict]

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


@app.post("/ml/v1/premium/calculate", response_model=PremiumV1Response)
def calculate_premium_v1(request: PremiumRequest):
    # Keep V1 endpoint stable while using the same placeholder model behavior.
    predicted_premium = 300.0
    confidence = 0.92

    # A simple bounded proxy for risk score used by demo clients.
    raw_risk = (request.zone_risk_rating * 0.5) + (request.disruption_frequency * 0.3) + (request.earnings_volatility * 0.2)
    risk_score = max(0.0, min(1.0, raw_risk))

    return PremiumV1Response(
        worker_id=request.worker_id,
        weekly_premium_inr=predicted_premium,
        risk_score=risk_score,
        confidence=confidence,
        shap_breakdown=[
            {"feature": "zone_risk_rating", "impact": round(request.zone_risk_rating * 0.5, 3)},
            {"feature": "disruption_frequency", "impact": round(request.disruption_frequency * 0.3, 3)},
            {"feature": "earnings_volatility", "impact": round(request.earnings_volatility * 0.2, 3)},
        ],
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
