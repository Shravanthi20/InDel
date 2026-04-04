from dataclasses import dataclass
from typing import Optional


@dataclass
class LoginRequest:
    password: str
    email: Optional[str] = None
    phone: Optional[str] = None


@dataclass
class RegisterRequest:
    username: str
    phone: str
    email: str
    password: str
    zone_level: Optional[str] = None
    zone_name: Optional[str] = None


@dataclass
class AuthResponse:
    token: str
    worker_id: str
    token_type: str = "Bearer"


@dataclass
class OTPRequest:
    phone: str


@dataclass
class OTPVerifyRequest:
    phone: str
    otp: str
