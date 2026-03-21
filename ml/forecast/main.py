from fastapi import FastAPI
from pydantic import BaseModel
from datetime import date

app = FastAPI(title="InDel Forecast Service")

class ForecastRequest(BaseModel):
    zone_id: int

class ForecastPoint(BaseModel):
    date: str
    disruption_probability: float

class ForecastResponse(BaseModel):
    zone_id: int
    forecast: list

@app.get("/health")
def health():
    return {"status": "ok", "service": "forecast-ml"}

@app.post("/forecast", response_model=ForecastResponse)
def generate_forecast(request: ForecastRequest):
    # Prophet time-series forecast
    forecast = [
        ForecastPoint(date="2026-03-22", disruption_probability=0.1),
        ForecastPoint(date="2026-03-23", disruption_probability=0.15),
        ForecastPoint(date="2026-03-24", disruption_probability=0.2),
        ForecastPoint(date="2026-03-25", disruption_probability=0.12),
        ForecastPoint(date="2026-03-26", disruption_probability=0.08),
        ForecastPoint(date="2026-03-27", disruption_probability=0.18),
        ForecastPoint(date="2026-03-28", disruption_probability=0.14),
    ]
    
    return ForecastResponse(
        zone_id=request.zone_id,
        forecast=forecast
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
