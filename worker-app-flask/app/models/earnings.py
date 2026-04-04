from dataclasses import dataclass, field
from typing import List


@dataclass
class EarningsSummary:
    this_week_actual: float = 0
    this_week_baseline: float = 0
    today_earnings: float = 0


@dataclass
class EarningsHistoryItem:
    date: str
    amount_inr: float
    order_count: int = 0


@dataclass
class EarningsHistoryResponse:
    history: List[EarningsHistoryItem] = field(default_factory=list)


@dataclass
class BaselineResponse:
    min_payout_inr: float = 0
    baseline_inr: float = 0
