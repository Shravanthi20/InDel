#!/usr/bin/env python3
"""
Batch ID Format Validator
Demonstrates the new batch ID format: ZONE + CITY_CODE + STATE_CODE + DATETIME
"""

import re
from datetime import datetime

BATCH_ID_PATTERN = r'^([A-C])([A-Z]{6}|[A-Z]{6})([A-Z]{4}|[A-Z]{4})(\d{14})$'

# For Zone B/C with different city/state counts
BATCH_ID_PATTERN_ZONE_B_C = r'^([BC])([A-Z]{6})([A-Z]{4})(\d{14})$'

def validate_batch_id(batch_id: str) -> bool:
    """Validate batch ID format"""
    # Zone A: 1 + 6 + 4 + 14 = 25 chars
    # Zone B: 1 + 6 + 4 + 14 = 25 chars
    # Zone C: 1 + 6 + 4 + 14 = 25 chars
    if len(batch_id) != 25:
        return False
    
    zone = batch_id[0]
    if zone not in ['A', 'B', 'C']:
        return False
    
    datetime_part = batch_id[-14:]
    try:
        datetime.strptime(datetime_part, '%Y%m%d%H%M%S')
    except ValueError:
        return False
    
    return True

def parse_batch_id(batch_id: str) -> dict:
    """Parse batch ID into components"""
    if not validate_batch_id(batch_id):
        return None
    
    zone = batch_id[0]
    datetime_part = batch_id[-14:]
    
    if zone == 'A':
        city_code = batch_id[1:7]  # 6 chars (single city)
        state_code = batch_id[7:11]  # 4 chars
        from_city = city_code
        to_city = city_code
        from_state = state_code
        to_state = state_code
    else:  # Zone B or C
        from_city_code = batch_id[1:4]  # 3 chars
        to_city_code = batch_id[4:7]  # 3 chars
        if zone == 'B':
            state_code = batch_id[7:11]  # 4 chars
            from_state = state_code
            to_state = state_code
        else:  # Zone C
            from_state = batch_id[7:9]  # 2 chars
            to_state = batch_id[9:11]  # 2 chars
    
    dt = datetime.strptime(datetime_part, '%Y%m%d%H%M%S')
    
    return {
        'zone': zone,
        'city_code': batch_id[1:7],
        'state_code': batch_id[7:11],
        'datetime': dt,
        'timestamp': datetime_part,
    }

def main():
    print("=" * 60)
    print("BATCH ID FORMAT VALIDATOR")
    print("=" * 60)
    print()
    
    # Test examples
    test_cases = [
        {
            'batch_id': 'ACHENNATAMI20260403123000',
            'description': 'Zone A: Chennai (same city), Tamil Nadu',
            'expected': True,
        },
        {
            'batch_id': 'BCHEBANTAMI20260403124500',
            'description': 'Zone B: Chennai→Bangalore, Tamil Nadu',
            'expected': True,
        },
        {
            'batch_id': 'CCHEMUMTAKA20260403130000',
            'description': 'Zone C: Chennai→Mumbai, Tamil Nadu→Karnataka',
            'expected': True,
        },
        {
            'batch_id': 'ACHENNATAMI2026040312',
            'description': 'Invalid: Incomplete datetime',
            'expected': False,
        },
        {
            'batch_id': 'ACHENNATAMI20260499999999',
            'description': 'Invalid: Invalid datetime (month 04, day 99)',
            'expected': False,
        },
    ]
    
    print("VALIDATION RESULTS:")
    print()
    
    for test in test_cases:
        batch_id = test['batch_id']
        is_valid = validate_batch_id(batch_id)
        status = "✓ VALID" if is_valid == test['expected'] else "✗ FAILED"
        
        print(f"{status} | {batch_id}")
        print(f"       Description: {test['description']}")
        
        if is_valid:
            parsed = parse_batch_id(batch_id)
            print(f"       Zone: {parsed['zone']}")
            print(f"       Datetime: {parsed['datetime'].strftime('%Y-%m-%d %H:%M:%S')}")
            print(f"       Timestamp: {parsed['timestamp']}")
        
        print()
    
    print("=" * 60)
    print("FORMAT SPECIFICATION")
    print("=" * 60)
    print()
    print("Zone A (Same City):")
    print("  Zone: A (1 char)")
    print("  City Code: 6 chars (from city name)")
    print("  State Code: 4 chars (from state name)")
    print("  Datetime: 14 chars (YYYYMMDDHHMMSS)")
    print("  Total: 25 chars")
    print()
    print("Zone B (Intra-State):")
    print("  Zone: B (1 char)")
    print("  City Code: 3+3 chars (from→to cities)")
    print("  State Code: 4 chars (state name)")
    print("  Datetime: 14 chars (YYYYMMDDHHMMSS)")
    print("  Total: 25 chars")
    print()
    print("Zone C (Inter-State):")
    print("  Zone: C (1 char)")
    print("  City Code: 3+3 chars (from→to cities)")
    print("  State Code: 2+2 chars (from→to states)")
    print("  Datetime: 14 chars (YYYYMMDDHHMMSS)")
    print("  Total: 25 chars")
    print()

if __name__ == '__main__':
    main()
