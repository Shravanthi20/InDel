import sys
import os
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv

# Ensure submodules can import their own files
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from premium.main import router as premium_router, load_model_instance as premium_load
from fraud.main import router as fraud_router
from forecast.main import router as forecast_router, train_all_zones as forecast_train

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("unified-ml")

def load_service_env():
    if os.getenv("INDEL_ENV", "").lower() == "production":
        log.info("Production mode detected, using process environment only")
        return

    base_dir = os.path.dirname(os.path.abspath(__file__))
    candidates = []
    explicit = os.getenv("ENV_FILE", "").strip()
    if explicit:
        candidates.append(explicit)
    candidates.extend([
        os.path.join(base_dir, ".env.local"),
        os.path.join(base_dir, ".env"),
    ])

    for candidate in candidates:
        if candidate and os.path.exists(candidate):
            load_dotenv(candidate)
            log.info("Loaded ML env file: %s", candidate)
            return

    log.info("No ML .env file found, using process environment")

def validate_env():
    missing = []
    for key in ["SERVICE_PORT"]:
        if not os.getenv(key, "").strip():
            missing.append(key)
    if missing:
        raise RuntimeError(f"Missing required ML environment variables: {', '.join(missing)}")

load_service_env()
validate_env()

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
    uvicorn.run(app, host="0.0.0.0", port=int(os.getenv("SERVICE_PORT", "8000")))
