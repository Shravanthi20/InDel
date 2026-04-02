import csv
import json
from collections import defaultdict
from itertools import combinations
from math import radians, sin, cos, sqrt, atan2

# Haversine distance in km
def haversine(lat1, lon1, lat2, lon2):
    R = 6371.0
    dlat = radians(lat2 - lat1)
    dlon = radians(lon2 - lon1)
    a = sin(dlat / 2) ** 2 + cos(radians(lat1)) * cos(radians(lat2)) * sin(dlon / 2) ** 2
    c = 2 * atan2(sqrt(a), sqrt(1 - a))
    return R * c

# Parse CSV and build city/state/lat/lon dicts
def parse_cities(file_path):
    cities = []
    states = defaultdict(list)
    city_coords = {}
    with open(file_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            state = row['State'].strip()
            city = row['Location'].split(' Latitude')[0].strip()
            lat = float(row['Latitude'])
            lon = float(row['Longitude'])
            cities.append((state, city))
            states[state].append(city)
            city_coords[city] = {'lat': lat, 'lon': lon}
    return cities, states, city_coords

cities, states, city_coords = parse_cities('Indian Cities Geo Data.csv')

# Zone B: city-to-city pairs in the same state, only if each is the nearest to the other
zone_b = []
for state, city_list in states.items():
    for city in city_list:
        lat1, lon1 = city_coords[city]['lat'], city_coords[city]['lon']
        # Find nearest city in the same state
        min_dist = float('inf')
        nearest = None
        for other in city_list:
            if city == other:
                continue
            lat2, lon2 = city_coords[other]['lat'], city_coords[other]['lon']
            dist = haversine(lat1, lon1, lat2, lon2)
            if dist < min_dist:
                min_dist = dist
                nearest = other
        # Only add the pair if this is the nearest neighbor
        if nearest:
            pair = tuple(sorted([city, nearest]))
            if not any(zb['from'] == pair[0] and zb['to'] == pair[1] for zb in zone_b):
                zone_b.append({
                    'from': pair[0],
                    'to': pair[1],
                    'state': state,
                    'distance_km': min_dist,
                    'from_lat': city_coords[pair[0]]['lat'],
                    'from_lon': city_coords[pair[0]]['lon'],
                    'to_lat': city_coords[pair[1]]['lat'],
                    'to_lon': city_coords[pair[1]]['lon']
                })

# Zone C: city-to-city pairs in different states, only if each is the nearest to the other (across states)
zone_c = []
for (state1, city1) in cities:
    lat1, lon1 = city_coords[city1]['lat'], city_coords[city1]['lon']
    min_dist = float('inf')
    nearest = None
    nearest_state = None
    for (state2, city2) in cities:
        if state1 == state2:
            continue
        lat2, lon2 = city_coords[city2]['lat'], city_coords[city2]['lon']
        dist = haversine(lat1, lon1, lat2, lon2)
        if dist < min_dist:
            min_dist = dist
            nearest = city2
            nearest_state = state2
    # Only add the pair if this is the nearest neighbor across states
    if nearest:
        pair = tuple(sorted([(state1, city1), (nearest_state, nearest)]))
        if not any(zc['from'] == pair[0][1] and zc['to'] == pair[1][1] for zc in zone_c):
            zone_c.append({
                'from': pair[0][1],
                'to': pair[1][1],
                'from_state': pair[0][0],
                'to_state': pair[1][0],
                'distance_km': min_dist,
                'from_lat': city_coords[pair[0][1]]['lat'],
                'from_lon': city_coords[pair[0][1]]['lon'],
                'to_lat': city_coords[pair[1][1]]['lat'],
                'to_lon': city_coords[pair[1][1]]['lon']
            })

# Store as JSON for API use
with open('zone_b.json', 'w', encoding='utf-8') as f:
    json.dump(zone_b, f, ensure_ascii=False, indent=2)
with open('zone_c.json', 'w', encoding='utf-8') as f:
    json.dump(zone_c, f, ensure_ascii=False, indent=2)

print('Filtered zone_b.json and zone_c.json generated (nearest city pairs only).')
