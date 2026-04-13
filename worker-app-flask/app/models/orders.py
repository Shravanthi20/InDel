from dataclasses import dataclass, field
from typing import List, Optional


@dataclass
class Order:
    order_id: str
    status: str
    earning_inr: float = 0
    customer_name: Optional[str] = None
    pickup_address: Optional[str] = None
    drop_address: Optional[str] = None


@dataclass
class OrderListResponse:
    orders: List[Order] = field(default_factory=list)


@dataclass
class BatchOrder:
    order_id: str
    status: str


@dataclass
class Batch:
    batch_id: str
    status: str
    order_count: int = 0
    orders: List[BatchOrder] = field(default_factory=list)


@dataclass
class BatchListResponse:
    batches: List[Batch] = field(default_factory=list)
