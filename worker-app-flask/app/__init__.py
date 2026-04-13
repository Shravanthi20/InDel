import os
from flask import Flask, redirect, session as flask_session, url_for
from app.config import Config


def create_app():
    base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    app = Flask(
        __name__,
        template_folder=os.path.join(base_dir, "templates"),
        static_folder=os.path.join(base_dir, "static"),
    )
    app.config.from_object(Config)

    @app.context_processor
    def inject_design_tokens():
        return {
            "brand_blue": "#00739D",
            "blue_deep": "#005A7A",
            "blue_soft": "#E6F3F7",
            "bg_warm_white": "#F5F7F9",
        }

    # Register blueprints
    from app.api.auth import auth_bp
    from app.api.worker import worker_bp
    from app.api.demo import demo_bp
    from app.api.platform import platform_bp

    app.register_blueprint(auth_bp)
    app.register_blueprint(worker_bp)
    app.register_blueprint(demo_bp)
    app.register_blueprint(platform_bp)

    @app.route("/")
    def index():
        if flask_session.get("token"):
            return redirect(url_for("worker.home"))
        return redirect(url_for("auth.login"))

    @app.errorhandler(404)
    def not_found(_error):
        return redirect(url_for("auth.login"))

    return app
