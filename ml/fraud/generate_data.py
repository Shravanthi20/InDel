import pandas as pd
import numpy as np
import random
import os

def generate_data(num_samples=500):
    np.random.seed(42)
    random.seed(42)
    
    data = []
    
    # Let's make sure the data directory exists
    os.makedirs('data', exist_ok=True)
    
    for i in range(1, num_samples + 1):
        claim_id = i
        worker_id = 100 + i
        
        # Decide if this is a fraudulent claim (about 15% fraud rate)
        is_fraud = random.random() < 0.15
        
        disruption_hours = random.choice([3, 4, 5, 6])
        
        if not is_fraud:
            # Genuine worker experiencing a real disruption
            baseline_hourly_income = round(random.uniform(90, 150), 2)
            actual_hourly_income = round(random.uniform(10, 40), 2)  # Significant drop due to disruption
            
            active_hours_during_disruption = disruption_hours
            orders_completed_during_disruption = random.randint(0, 3) 
            
            gps_zone_match = 1
            distance_from_zone_center = round(random.uniform(0.5, 3.0), 2)
            
            total_active_weeks = random.randint(4, 52)
            past_claim_count = random.randint(0, 2)
            approved_claims = past_claim_count  # Usually all are approved for good workers
            
            zone_risk_score = round(random.uniform(0.4, 0.9), 2)
            disruption_type = random.choice(['weather', 'aqi', 'curfew'])
            
        else:
            # Fraudulent worker profiles
            fraud_type = random.choice(['zone_hopper', 'spoofer', 'idle'])
            
            baseline_hourly_income = round(random.uniform(90, 150), 2)
            
            if fraud_type == 'idle':
                # Hanging around making zero deliveries but claiming money anyway
                actual_hourly_income = 0
                active_hours_during_disruption = random.randint(1, disruption_hours)
                orders_completed_during_disruption = 0
                gps_zone_match = 1
                distance_from_zone_center = round(random.uniform(0.1, 2.0), 2)
                
            elif fraud_type == 'zone_hopper':
                # Way outside the declared zone where the disruption actually is
                actual_hourly_income = round(random.uniform(10, 30), 2)
                active_hours_during_disruption = disruption_hours
                orders_completed_during_disruption = random.randint(1, 2)
                gps_zone_match = 0  # GPS doesn't match the zone they claim to be in!
                distance_from_zone_center = round(random.uniform(10.0, 30.0), 2)
                
            elif fraud_type == 'spoofer':
                # Actually completing a lot of orders but submitting a claim anyway
                actual_hourly_income = round(random.uniform(70, 110), 2) 
                active_hours_during_disruption = disruption_hours
                orders_completed_during_disruption = random.randint(4, 8)
                gps_zone_match = 1
                distance_from_zone_center = round(random.uniform(0.0, 0.5), 2)
            
            total_active_weeks = random.randint(1, 10)
            past_claim_count = random.randint(2, 6) # High claim count
            approved_claims = random.randint(0, max(0, past_claim_count - 2))
            
            zone_risk_score = round(random.uniform(0.1, 0.5), 2)  
            disruption_type = random.choice(['weather', 'aqi'])
        
        # Calculate derived fields based on user formulas
        earnings_drop_ratio = max(0.0, round((baseline_hourly_income - actual_hourly_income) / baseline_hourly_income, 4))
        
        avg_orders_per_hour = round(orders_completed_during_disruption / active_hours_during_disruption, 4) if active_hours_during_disruption > 0 else 0.0
        
        claim_frequency = round(past_claim_count / total_active_weeks, 4) if total_active_weeks > 0 else 0.0
        
        approval_ratio = round(approved_claims / past_claim_count, 4) if past_claim_count > 0 else 1.0
        
        data.append([
            claim_id, worker_id, baseline_hourly_income, actual_hourly_income, 
            earnings_drop_ratio, disruption_hours, active_hours_during_disruption, 
            orders_completed_during_disruption, avg_orders_per_hour, gps_zone_match, 
            distance_from_zone_center, past_claim_count, claim_frequency, 
            approval_ratio, zone_risk_score, disruption_type
        ])
        
    df = pd.DataFrame(data, columns=[
        'claim_id', 'worker_id', 'baseline_hourly_income', 'actual_hourly_income',
        'earnings_drop_ratio', 'disruption_hours', 'active_hours_during_disruption',
        'orders_completed_during_disruption', 'avg_orders_per_hour', 'gps_zone_match',
        'distance_from_zone_center', 'past_claim_count', 'claim_frequency',
        'approval_ratio', 'zone_risk_score', 'disruption_type'
    ])
    
    # Save the dataframe
    output_path = 'data/synthetic_claim_patterns.csv'
    df.to_csv(output_path, index=False)
    print(f"Generated {num_samples} records and saved to {output_path}")

if __name__ == "__main__":
    generate_data(1000)
