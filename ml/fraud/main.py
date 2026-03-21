from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="InDel Fraud Detection Service")

class FraudRequest(BaseModel):
    claim_id: int
    worker_id: int
    claim_amount: float
    disruption_type: str
    worker_history: dict

class FraudResponse(BaseModel):
    fraud_score: float
    verdict: str
    signals: list
    confidence: float

@app.get("/health")
def health():
    return {"status": "ok", "service": "fraud-ml"}

@app.post("/score", response_model=FraudResponse)
def score_claim(request: FraudRequest):
    # 3-layer fraud detection
    # Layer 1: Isolation Forest
    # Layer 2: DBSCAN clustering
    # Layer 3: Rule-based checks
    
    fraud_score = 0.2  # Placeholder
    verdict = "clean"
    signals = []
    confidence = 0.85
    
    return FraudResponse(
        fraud_score=fraud_score,
        verdict=verdict,
        signals=signals,
        confidence=confidence
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
