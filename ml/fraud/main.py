from fastapi import FastAPI
from pydantic import BaseModel, Field
from typing import List, Optional
import uuid
from datetime import datetime

from scorer import FraudScorer

from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(title="InDel Fraud Detection Service — 3-Layer Stack")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
scorer = FraudScorer()

# ─── Schemas ────────────────────────────────────────────────────────────────

class WorkerHistory(BaseModel):
    total_claims_last_8_weeks: int = 0
    approved_claims_last_8_weeks: int = 0
    avg_claim_amount: float = 0.0
    earnings_variance: float = 0.0
    zone_change_count: int = 0
    days_active: int = 30
    delivery_attempt_rate: float = 0.8   # attempts / available orders

class FraudRequest(BaseModel):
    claim_id: int
    worker_id: int
    zone_id: int
    claim_amount: float
    baseline_earnings: float
    disruption_type: str
    disruption_hours: float = 4.0
    gps_in_zone: bool = True
    distance_from_zone_center: float = 0.5
    deliveries_during_disruption: int = 0
    zone_avg_claim_amount: Optional[float] = None
    zone_risk_score: float = 0.5
    worker_history: WorkerHistory = Field(default_factory=WorkerHistory)

class FraudSignal(BaseModel):
    name: str
    impact: float
    description: str

class FraudResponse(BaseModel):
    claim_id: int
    fraud_score: float          # 0.0 = clean, 1.0 = definite fraud
    verdict: str                # "safe" | "review" | "flagged"
    signals: List[FraudSignal]
    confidence: float
    routing: str                # "auto_approve" | "manual_review" | "auto_reject"
    request_id: str
    timestamp: datetime

# ─── Health ─────────────────────────────────────────────────────────────────

@app.get("/health")
def health():
    return {
        "status": "ok", 
        "service": "fraud-ml", 
        "layers": ["isolation_forest", "dbscan", "rules"], 
        "models_loaded": scorer.models_loaded
    }

# ─── Score endpoint ──────────────────────────────────────────────────────────

@app.post("/ml/v1/fraud/score", response_model=FraudResponse)
def score_claim(request: FraudRequest):
    result = scorer.score(request)
    return FraudResponse(
        claim_id=request.claim_id,
        fraud_score=result["score"],
        verdict=result["verdict"],
        signals=[FraudSignal(**s) for s in result["signals"]],
        confidence=result["confidence"],
        routing=result["routing"],
        request_id=f"fraud_{uuid.uuid4().hex[:8]}",
        timestamp=datetime.utcnow()
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
