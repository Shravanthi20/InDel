import unittest
from scripts.fake_order_publisher import random_order, CITY_STATE_LOOKUP, determine_zone_and_vehicle


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
        order = random_order(1)
        self.assertIn(order['zone_type'], ['intra-zone', 'inter-state', 'unknown'])
        self.assertIn(order['required_vehicle_type'], ['bike/small van', 'van/truck', 'unknown'])
        self.assertIn('needs_hub_transfer', order)

if __name__ == "__main__":
    unittest.main()
