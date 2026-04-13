import csv
from typing import Dict

def load_city_state_lookup(csv_path: str) -> Dict[str, str]:
    """
    Loads a mapping from area name (city in CSV, stripped) to city name (state in CSV).
    Args:
        csv_path: Path to the Indian Cities Geo Data CSV file.
    Returns:
        Dictionary mapping area name (city, e.g. 'Port Blair') to city name (state).
    """
    lookup = {}
    with open(csv_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            area = row['Location'].replace(' Latitude and Longitude', '').strip()
            city = row['State'].strip()
            lookup[area] = city
    return lookup

# --- Zone Rule Engine ---
def determine_zone_and_vehicle(source_city: str, dest_city: str, city_state_lookup: dict) -> dict:
    """
    Determines the zone type and required vehicle for an order.
    Args:
        source_city: Name of the source city.
        dest_city: Name of the destination city.
        city_state_lookup: Dict mapping city name to state name.
    Returns:
        Dict with zone_type, required_vehicle_type, needs_hub_transfer.
    """
    source_state = city_state_lookup.get(source_city)
    dest_state = city_state_lookup.get(dest_city)
    if not source_state or not dest_state:
        return {
            'zone_type': 'unknown',
            'required_vehicle_type': 'unknown',
            'needs_hub_transfer': False
        }
    if source_state == dest_state:
        return {
            'zone_type': 'intra-zone',
            'required_vehicle_type': 'bike/small van',
            'needs_hub_transfer': False
        }
    else:
        # For now, assume all inter-state orders are possible directly
        return {
            'zone_type': 'inter-state',
            'required_vehicle_type': 'van/truck',
            'needs_hub_transfer': False  # Set to True if you want to force hub transfer for some cases
        }

if __name__ == "__main__":
    # Example usage
    lookup = load_city_state_lookup("../Indian Cities Geo Data.csv")
    print(f"Total cities loaded: {len(lookup)}")
    print("Sample:", list(lookup.items())[:5])
    print(determine_zone_and_vehicle("Chennai", "Bangalore", lookup))
    print(determine_zone_and_vehicle("Chennai", "Chennai", lookup))
