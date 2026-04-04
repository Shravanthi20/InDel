"""
Zone levels and names — mirrors Kotlin WorkerSelectionOptions.kt
"""

# ── Zone Levels ──────────────────────────────────────────────────────────────
ZONE_LEVELS = [
    {"level": "A", "label": "A local"},
    {"level": "B", "label": "B intra-state"},
    {"level": "C", "label": "C metro-to-metro"},
]

# ── Zone Names by Level ──────────────────────────────────────────────────────
ZONE_NAMES_LEVEL_A = ["Tambaram", "Chromepet", "Pallavaram", "Selaiyur", "Velachery", "Adyar"]
ZONE_NAMES_LEVEL_B = ["Chennai Central", "Madurai", "Coimbatore", "Trichy", "Salem"]
ZONE_NAMES_LEVEL_C = ["Chennai-Bangalore", "Chennai-Hyderabad", "Chennai-Mumbai", "Chennai-Delhi"]

def get_zone_names_for_level(level: str) -> list:
    """Get zone names for a given level (A, B, or C)"""
    level = level.strip().upper()
    if level == "A":
        return ZONE_NAMES_LEVEL_A
    elif level == "B":
        return ZONE_NAMES_LEVEL_B
    elif level == "C":
        return ZONE_NAMES_LEVEL_C
    return []

# ── Vehicles ─────────────────────────────────────────────────────────────────
ALL_VEHICLES = [
    "scooter",
    "motorcycle",
    "auto-rickshaw",
    "car",
    "van",
    "mini-truck",
    "truck",
]

FOUR_WHEELER_VEHICLES = [
    "car",
    "van",
    "mini-truck",
    "truck",
]

def get_vehicles_for_level(level: str) -> list:
    """Get allowed vehicles for a zone level"""
    level = level.strip().upper()
    if level == "C":
        return FOUR_WHEELER_VEHICLES
    return ALL_VEHICLES

def is_valid_upi(upi_id: str) -> bool:
    """Validate UPI ID format"""
    trimmed = upi_id.strip()
    return (
        "@" in trimmed
        and not trimmed.startswith("@")
        and not trimmed.endswith("@")
        and trimmed.count("@") == 1
    )
