"""
Session store — mirrors Kotlin PreferencesDataStore.
Provides login_required decorator and helpers to read/write session-persisted auth.
"""
from functools import wraps
from flask import session as flask_session, redirect, url_for


def login_required(f):
    """Guard decorator — clears session and redirects to login on missing token (mirrors 401 behaviour)."""
    @wraps(f)
    def wrapper(*args, **kwargs):
        if not flask_session.get("token"):
            flask_session.clear()
            return redirect(url_for("auth.login"))
        return f(*args, **kwargs)
    return wrapper


def save_auth(token: str, worker_id: str, token_type: str = "Bearer"):
    flask_session["token"] = token
    flask_session["worker_id"] = worker_id
    flask_session["token_type"] = token_type


def clear_auth():
    flask_session.clear()


def get_worker_id() -> str:
    return flask_session.get("worker_id", "")


def get_token() -> str:
    return flask_session.get("token", "")
