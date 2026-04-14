import csv
import json
from collections import defaultdict
from itertools import combinations

# Parse Indian Cities Geo Data.csv and generate zone data

def parse_cities(file_path):
    cities = []
    states = defaultdict(list)
    with open(file_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            state = row['State'].strip()
            location = row['Location'].split(' Latitude')[0].strip()
            lat = float(row['Latitude']) if row['Latitude'] else None
            lon = float(row['Longitude']) if row['Longitude'] else None
            city_obj = {
                'name': location,
                'state': state,
                'latitude': lat,
                'longitude': lon
            }
            cities.append(city_obj)
            states[state].append(city_obj)
    return cities, states

cities, states = parse_cities(r'C:\Users\gayat\projects\get_into\InDel\Indian Cities Geo Data.csv')

# Zone A: unique city objects (first 15)
zone_a = []
seen = set()
for city in cities:
    key = (city['name'], city['state'])
    if key not in seen:
        zone_a.append(city)
        seen.add(key)
    if len(zone_a) == 15:
        break

# Zone B: city-to-city pairs in the same state (first 15)
zone_b = []
for state, city_list in states.items():
    city_list = [c for c in city_list if c['latitude'] is not None and c['longitude'] is not None]
    for c1, c2 in combinations(city_list, 2):
        zone_b.append({
            'from': c1['name'],
            'to': c2['name'],
            'state': state,
            'from_lat': c1['latitude'],
            'from_lon': c1['longitude'],
            'to_lat': c2['latitude'],
            'to_lon': c2['longitude']
        })
        if len(zone_b) == 15:
            break
    if len(zone_b) == 15:
        break

# Zone C: city-to-city pairs in different states (first 15)
zone_c = []
for c1, c2 in combinations(cities, 2):
    if c1['state'] != c2['state'] and c1['latitude'] is not None and c1['longitude'] is not None and c2['latitude'] is not None and c2['longitude'] is not None:
        zone_c.append({
            'from': c1['name'],
            'to': c2['name'],
            'from_state': c1['state'],
            'to_state': c2['state'],
            'from_lat': c1['latitude'],
            'from_lon': c1['longitude'],
            'to_lat': c2['latitude'],
            'to_lon': c2['longitude']
        })
        if len(zone_c) == 15:
            break

# Store as JSON for API use
with open('zone_a.json', 'w', encoding='utf-8') as f:
    json.dump(zone_a, f, ensure_ascii=False, indent=2)
with open('zone_b.json', 'w', encoding='utf-8') as f:
    json.dump(zone_b, f, ensure_ascii=False, indent=2)
with open('zone_c.json', 'w', encoding='utf-8') as f:
    json.dump(zone_c, f, ensure_ascii=False, indent=2)

print('Zone data generated and saved as zone_a.json, zone_b.json, zone_c.json')
