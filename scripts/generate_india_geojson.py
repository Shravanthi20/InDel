#!/usr/bin/env python3
"""
Generate india.geojson from zone_a.json for all cities as Point features.
Output: platform-dashboard/public/india.geojson
"""
import json
import os

ZONE_A_PATH = os.path.join(os.path.dirname(__file__), "..", "zone_a.json")
OUTPUT_PATH = os.path.join(os.path.dirname(__file__), "..", "platform-dashboard", "public", "india.geojson")

with open(ZONE_A_PATH, encoding="utf-8") as f:
    cities = json.load(f)

features = []
for city in cities:
    # Support both dict and legacy string format
    if isinstance(city, dict):
        name = city.get("city") or city.get("name")
        state = city.get("state")
        lat = city.get("lat")
        lng = city.get("long") or city.get("lng")
    else:
        # Legacy: just city name, skip
        continue
    if not (name and state and lat is not None and lng is not None):
        continue
    features.append({
        "type": "Feature",
        "properties": {
            "city": name,
            "state": state,
            "lat": lat,
            "long": lng
        },
        "geometry": {
            "type": "Point",
            "coordinates": [lng, lat]
        }
    })

geojson = {
    "type": "FeatureCollection",
    "features": features
}

with open(OUTPUT_PATH, "w", encoding="utf-8") as f:
    json.dump(geojson, f, ensure_ascii=False, indent=2)

print(f"Wrote {len(features)} features to {OUTPUT_PATH}")
