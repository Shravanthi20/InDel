"""
Generate 90 days of realistic synthetic disruption history for all 4 zones.
Uses only Python standard library — no dependencies needed.
Run: python generate_history.py
"""
import csv
import random
import math
from datetime import date, timedelta

random.seed(42)

OUTPUT = "synthetic_zone_history.csv"
ZONES = [1, 2, 3, 4]
DAYS = 90

# Zone characteristics (baseline disruption rate, temp range, aqi range, order_volume range)
ZONE_CONFIG = {
    1: {"base_rate": 0.35, "temp": (26, 38), "aqi": (55, 135), "vol": (800, 1100), "label": "urban_high"},
    2: {"base_rate": 0.15, "temp": (22, 33), "aqi": (40, 90),  "vol": (500,  750), "label": "suburban_low"},
    3: {"base_rate": 0.25, "temp": (24, 36), "aqi": (50, 115), "vol": (650,  950), "label": "mixed_med"},
    4: {"base_rate": 0.08, "temp": (20, 30), "aqi": (30, 70),  "vol": (350,  550), "label": "rural_low"},
}

end_date   = date(2026, 4, 14)
start_date = end_date - timedelta(days=DAYS - 1)

rows = []
for zone_id, cfg in ZONE_CONFIG.items():
    for i in range(DAYS):
        d = start_date + timedelta(days=i)

        # Weekday effect — more disruptions mid-week (Tue-Thu) and weekends
        weekday = d.weekday()  # Mon=0, Sun=6
        weekday_factor = 1.0
        if weekday in (1, 2, 3):   # Tue–Thu: peak delivery
            weekday_factor = 1.4
        elif weekday in (5, 6):    # Sat–Sun: lower order baseline, higher weather risk
            weekday_factor = 0.75

        # Seasonal trend — slightly higher disruption in Jan–Feb (winter) and Mar–Apr (pre-summer heat)
        month_factor = 1.0
        if d.month in (1, 2):
            month_factor = 1.25
        elif d.month == 3:
            month_factor = 1.15
        elif d.month == 4:
            month_factor = 1.10

        # Progressive trend — disruptions increasing slightly over time (system growth)
        progress = i / DAYS
        trend_factor = 1.0 + 0.20 * progress

        # Temperature
        t_min, t_max = cfg["temp"]
        temp = round(random.uniform(t_min, t_max) + math.sin(i / 14.0) * 3, 1)

        # AQI — correlated with temperature
        a_min, a_max = cfg["aqi"]
        aqi = int(random.uniform(a_min, a_max) + (temp - t_min) * 0.6)
        aqi = max(a_min, min(a_max + 30, aqi))

        # AQI spike factor — higher AQI → more disruptions
        aqi_factor = 1.0 + max(0, (aqi - 80) / 80)

        # Order volume — inversely correlated with disruptions
        v_min, v_max = cfg["vol"]
        order_volume = int(random.uniform(v_min, v_max) * (1.0 - 0.15 * progress))

        # Final disruption probability
        prob = cfg["base_rate"] * weekday_factor * month_factor * trend_factor * aqi_factor
        prob = min(prob, 0.85)

        # Sample disruption count from Poisson-like distribution
        events = 0
        for _ in range(5):  # max 5 events per day per zone
            if random.random() < prob:
                events += 1

        rows.append({
            "date":              d.isoformat(),
            "zone_id":           zone_id,
            "disruption_events": events,
            "avg_temperature":   temp,
            "aqi":               aqi,
            "order_volume":      order_volume,
        })

with open(OUTPUT, "w", newline="") as f:
    writer = csv.DictWriter(f, fieldnames=["date", "zone_id", "disruption_events", "avg_temperature", "aqi", "order_volume"])
    writer.writeheader()
    writer.writerows(rows)

total = sum(r["disruption_events"] for r in rows)
print(f"Generated {len(rows)} rows ({DAYS} days × {len(ZONES)} zones)")
print(f"Total disruption events across all zones: {total}")
for z in ZONES:
    zone_rows = [r for r in rows if r["zone_id"] == z]
    z_total = sum(r["disruption_events"] for r in zone_rows)
    print(f"  Zone {z}: {z_total} events over {DAYS} days ({z_total/DAYS:.2f}/day avg)")
print(f"Written to {OUTPUT}")
