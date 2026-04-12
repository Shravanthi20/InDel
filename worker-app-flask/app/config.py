import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    SECRET_KEY = os.environ.get("SECRET_KEY", "dev-secret-change-me")
    API_BASE_URL = os.environ.get("API_BASE_URL", "http://127.0.0.1:8004").rstrip("/")
    API_TIMEOUT_SECONDS = float(os.environ.get("API_TIMEOUT_SECONDS", "12"))
    DEBUG = os.environ.get("FLASK_DEBUG", "true").lower() == "true"
    SESSION_COOKIE_HTTPONLY = True
    SESSION_COOKIE_SAMESITE = os.environ.get("SESSION_COOKIE_SAMESITE", "Lax")
    SESSION_COOKIE_SECURE = os.environ.get("SESSION_COOKIE_SECURE", "false").lower() == "true"
