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
            city = row['Location'].split(' Latitude')[0].strip()
            cities.append((state, city))
            states[state].append(city)
    return cities, states

cities, states = parse_cities('Indian Cities Geo Data.csv')

# Zone A: unique city names
zone_a = sorted(set(city for _, city in cities))

# Zone B: city-to-city pairs in the same state
zone_b = []
for state, city_list in states.items():
    for c1, c2 in combinations(sorted(set(city_list)), 2):
        zone_b.append({'from': c1, 'to': c2, 'state': state})

# Zone C: city-to-city pairs in different states
zone_c = []
for (state1, city1), (state2, city2) in combinations(sorted(set(cities)), 2):
    if state1 != state2:
        zone_c.append({'from': city1, 'to': city2, 'from_state': state1, 'to_state': state2})

# Store as JSON for API use
with open('zone_a.json', 'w', encoding='utf-8') as f:
    json.dump(zone_a, f, ensure_ascii=False, indent=2)
with open('zone_b.json', 'w', encoding='utf-8') as f:
    json.dump(zone_b, f, ensure_ascii=False, indent=2)
with open('zone_c.json', 'w', encoding='utf-8') as f:
    json.dump(zone_c, f, ensure_ascii=False, indent=2)

print('Zone data generated and saved as zone_a.json, zone_b.json, zone_c.json')
