from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass
class SimpleMessageResponse:
    message: str


@dataclass
class SessionResponse:
    session_id: str
    status: str
    started_at: Optional[str] = None
    ended_at: Optional[str] = None


@dataclass
class FraudSignal:
    signal_id: str
    level: str
    description: str


@dataclass
class Notification:
    notification_id: str
    title: str
    body: str
    read: bool = False


@dataclass
class ZoneConfig:
    zone_level: str
    zone_name: str
    config: Dict[str, str] = field(default_factory=dict)
