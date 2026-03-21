#!/usr/bin/env python3
"""
Generate synthetic data for InDel demo database.
Creates 4 weeks of fake order history and earnings data.
"""

import random
import datetime
import psycopg2
from psycopg2.extras import execute_batch
from dotenv import load_dotenv
import os

load_dotenv()

DB_HOST = os.getenv('DB_HOST', 'localhost')
DB_PORT = os.getenv('DB_PORT', '5432')
DB_USER = os.getenv('DB_USER', 'indel')
DB_PASSWORD = os.getenv('DB_PASSWORD', 'password')
DB_NAME = os.getenv('DB_NAME', 'indel_demo')

def connect_db():
    """Connect to PostgreSQL database."""
    return psycopg2.connect(
        host=DB_HOST,
        port=DB_PORT,
        user=DB_USER,
        password=DB_PASSWORD,
        database=DB_NAME
    )

def generate_orders(conn, num_orders=500):
    """Generate sample orders for past 4 weeks."""
    cursor = conn.cursor()
    orders = []
    
    for _ in range(num_orders):
        worker_id = random.randint(1, 4)
        zone_id = random.randint(1, 4)
        order_value = round(random.uniform(300, 1000), 2)
        days_ago = random.randint(0, 28)
        created_at = datetime.datetime.now() - datetime.timedelta(days=days_ago)
        
        orders.append((worker_id, zone_id, order_value, created_at))
    
    execute_batch(cursor, 
        "INSERT INTO orders (worker_id, zone_id, order_value, created_at) VALUES (%s, %s, %s, %s)",
        orders)
    
    conn.commit()
    print(f"Generated {num_orders} orders")

def generate_earnings(conn):
    """Generate earnings records for past 4 weeks."""
    cursor = conn.cursor()
    earnings = []
    
    for worker_id in range(1, 5):
        for days_ago in range(28, 0, -1):
            date = (datetime.datetime.now() - datetime.timedelta(days=days_ago)).date()
            
            # Random 30% chance of disruption (0 earnings)
            if random.random() < 0.3:
                hours_worked = 0
                amount_earned = 0
            else:
                hours_worked = random.randint(6, 12)
                amount_earned = round(random.uniform(3000, 8000), 2)
            
            earnings.append((worker_id, date, hours_worked, amount_earned))
    
    execute_batch(cursor,
        "INSERT INTO earnings_records (worker_id, date, hours_worked, amount_earned) VALUES (%s, %s, %s, %s)",
        earnings)
    
    conn.commit()
    print(f"Generated {len(earnings)} earnings records")

def generate_disruptions(conn):
    """Generate sample disruptions."""
    cursor = conn.cursor()
    disruptions = [
        (1, 'weather', 'high', datetime.datetime.now() - datetime.timedelta(days=7)),
        (1, 'aqi', 'medium', datetime.datetime.now() - datetime.timedelta(days=5)),
        (2, 'order_drop', 'high', datetime.datetime.now() - datetime.timedelta(days=3)),
        (3, 'weather', 'medium', datetime.datetime.now() - datetime.timedelta(days=2)),
        (4, 'aqi', 'low', datetime.datetime.now() - datetime.timedelta(days=1)),
    ]
    
    execute_batch(cursor,
        "INSERT INTO disruptions (zone_id, type, severity, signal_timestamp) VALUES (%s, %s, %s, %s)",
        disruptions)
    
    conn.commit()
    print(f"Generated {len(disruptions)} disruptions")

def generate_claims(conn):
    """Generate sample claims."""
    cursor = conn.cursor()
    claims = [
        (1, 1, 5000, 'approved'),
        (2, 2, 1500, 'approved'),
        (3, 3, 3000, 'pending'),
        (4, 4, 2000, 'denied'),
    ]
    
    execute_batch(cursor,
        "INSERT INTO claims (disruption_id, worker_id, claim_amount, status) VALUES (%s, %s, %s, %s)",
        claims)
    
    conn.commit()
    print(f"Generated {len(claims)} claims")

def main():
    print("Generating synthetic data for InDel demo...")
    
    conn = connect_db()
    
    try:
        generate_orders(conn, num_orders=500)
        generate_earnings(conn)
        generate_disruptions(conn)
        generate_claims(conn)
        
        print("\nSynthetic data generation complete!")
    
    finally:
        conn.close()

if __name__ == '__main__':
    main()
