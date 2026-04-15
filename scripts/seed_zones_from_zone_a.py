#!/usr/bin/env python3
"""
Seed the backend DB's zones table with all cities/zones from zone_a.json.
Usage: python scripts/seed_zones_from_zone_a.py
"""
import json
import psycopg2
import os

ZONE_A_PATH = os.path.join(os.path.dirname(__file__), "..", "zone_a.json")

# DB connection details (edit as needed or use env vars)
DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", 5432))
DB_USER = os.environ.get("DB_USER", "indel")
DB_PASSWORD = os.environ.get("DB_PASSWORD", "demo_password")
DB_NAME = os.environ.get("DB_NAME", "indel_demo")

with open(ZONE_A_PATH, encoding="utf-8") as f:
    cities = json.load(f)

conn = psycopg2.connect(
    host=DB_HOST,
    port=DB_PORT,
    user=DB_USER,
    password=DB_PASSWORD,
    dbname=DB_NAME
)
cur = conn.cursor()

for city in cities:
    name = city.get("city") or city.get("name")
    state = city.get("state")
    lat = city.get("lat")
    lng = city.get("long") or city.get("lng")
    if not (name and state and lat is not None and lng is not None):
        continue
    # Upsert zone by name+state
    cur.execute("""
        INSERT INTO zones (name, city, state, lat, long, level, risk_rating)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (name, state) DO UPDATE SET
            city=EXCLUDED.city,
            lat=EXCLUDED.lat,
            long=EXCLUDED.long,
            level=EXCLUDED.level,
            risk_rating=EXCLUDED.risk_rating
    """, (name, name, state, lat, lng, 'A', 0.5))

conn.commit()
cur.close()
conn.close()
print(f"Seeded {len(cities)} zones to DB.")
