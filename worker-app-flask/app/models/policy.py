from dataclasses import dataclass
from typing import Optional


@dataclass
class Policy:
    status: str
    plan_name: Optional[str] = None
    coverage_ratio: float = 0
    weekly_premium_inr: int = 0


@dataclass
class PolicyResponse:
    policy: Policy


@dataclass
class PremiumResponse:
    amount_inr: int
    due_date: Optional[str] = None


@dataclass
class PayPremiumRequest:
    amount: Optional[int] = None
