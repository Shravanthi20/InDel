from dataclasses import dataclass
from typing import Optional


@dataclass
class WorkerProfile:
    worker_id: str
    name: str
    vehicle_type: str
    upi_id: str
    zone_level: Optional[str] = None
    zone_name: Optional[str] = None
    city: Optional[str] = None
    from_city: Optional[str] = None
    to_city: Optional[str] = None
    coverage_status: Optional[str] = None


@dataclass
class OnboardRequest:
    name: str
    vehicle_type: str
    upi_id: str
    zone_level: Optional[str] = None
    zone_name: Optional[str] = None
    zone_id: Optional[int] = None
    city: Optional[str] = None
    from_city: Optional[str] = None
    to_city: Optional[str] = None
    vehicle_name: Optional[str] = None


@dataclass
class OnboardResponse:
    message: str
    worker_id: str
