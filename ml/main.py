import sys
import os
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware

# Ensure submodules can import their own files
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from premium.main import router as premium_router, load_model_instance as premium_load
from fraud.main import router as fraud_router
from forecast.main import router as forecast_router, train_all_zones as forecast_train

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("unified-ml")

# Global status tracking
initialization_status = {
    "premium": "pending",
    "forecast": "pending"
}

def run_initialization():
    log.info("Starting background initialization...")
    try:
        premium_load()
        initialization_status["premium"] = "ready"
    except Exception as e:
        log.error(f"Premium initialization failed: {e}")
        initialization_status["premium"] = "failed"
        
    try:
        forecast_train()
        initialization_status["forecast"] = "ready"
    except Exception as e:
        log.error(f"Forecast initialization failed: {e}")
        initialization_status["forecast"] = "failed"
    log.info("Background initialization complete.")

@asynccontextmanager
async def lifespan(app: FastAPI):
    log.info("Unified ML Service starting...")
    # Initialize in background to let health checks pass immediately
    from threading import Thread
    thread = Thread(target=run_initialization)
    thread.start()
    yield
    log.info("Shutting down Unified ML Service.")

app = FastAPI(title="InDel Unified ML Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=False,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(premium_router)
app.include_router(fraud_router)
app.include_router(forecast_router)

@app.get("/health")
def health():
    return {
        "status": "ok",
        "service": "unified-ml",
        "initialization": initialization_status,
        "components": ["premium", "fraud", "forecast"]
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
