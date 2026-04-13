from dataclasses import dataclass, field
from typing import List, Optional


@dataclass
class ClaimFactor:
    key: str
    value: str


@dataclass
class DisruptionWindow:
    start_at: str
    end_at: str


@dataclass
class Claim:
    claim_id: str
    status: str
    requested_amount_inr: float = 0
    approved_amount_inr: float = 0
    reason: Optional[str] = None
    factors: List[ClaimFactor] = field(default_factory=list)


@dataclass
class WalletResponse:
    balance_inr: float = 0
    pending_inr: float = 0


@dataclass
class Payout:
    payout_id: str
    amount_inr: float
    status: str
    created_at: str
