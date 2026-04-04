from dataclasses import dataclass, field
from typing import List, Optional


@dataclass
class DeliveryPlan:
    plan_id: str
    plan_name: str
    range_start: int
    range_end: int
    weekly_premium_inr: int
    coverage_ratio: float
    max_payout_inr: int
    description: Optional[str] = None


@dataclass
class PlanListResponse:
    plans: List[DeliveryPlan] = field(default_factory=list)


@dataclass
class PlanSelectionRequest:
    plan_id: str
    payment_amount_inr: int
    expected_deliveries: Optional[int] = None
    payment_confirmed: bool = True


@dataclass
class PlanSelectionResponse:
    message: str
    policy_status: Optional[str] = None
