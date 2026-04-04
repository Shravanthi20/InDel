import unittest
from scripts.fake_order_publisher import (
    CITY_STATE_LOOKUP,
    determine_zone_and_vehicle,
    load_zone_pairs,
    random_order_from_pair,
)


class TestOrderEligibility(unittest.TestCase):
    def test_intra_zone(self):
        # Use two areas in the same city (e.g., 'Port Blair' to 'Port Blair')
        result = determine_zone_and_vehicle('Port Blair', 'Port Blair', CITY_STATE_LOOKUP)
        self.assertEqual(result['zone_type'], 'intra-zone')
        self.assertEqual(result['required_vehicle_type'], 'bike/small van')
        self.assertFalse(result['needs_hub_transfer'])

    def test_inter_state(self):
        # Use two areas in different cities (e.g., 'Port Blair' to 'Addanki')
        result = determine_zone_and_vehicle('Port Blair', 'Addanki', CITY_STATE_LOOKUP)
        self.assertEqual(result['zone_type'], 'inter-state')
        self.assertEqual(result['required_vehicle_type'], 'van/truck')
        self.assertFalse(result['needs_hub_transfer'])

    def test_random_order_fields(self):
        zone_a, _, _ = load_zone_pairs()
        self.assertGreater(len(zone_a), 0)
        pair = {
            "from": "Port Blair",
            "to": "Port Blair",
            "from_state": CITY_STATE_LOOKUP.get("Port Blair", "Andaman and Nicobar Islands"),
            "to_state": CITY_STATE_LOOKUP.get("Port Blair", "Andaman and Nicobar Islands"),
            "distance_km": 8.0,
        }
        order = random_order_from_pair(1, pair, "same-city", zone_id=1)
        self.assertIn(order['zone_type'], ['intra-zone', 'inter-state', 'unknown'])
        self.assertIn(order['required_vehicle_type'], ['bike/small van', 'van/truck', 'unknown'])
        self.assertIn('needs_hub_transfer', order)

if __name__ == "__main__":
    unittest.main()
