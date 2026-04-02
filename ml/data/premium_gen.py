import pandas as pd
import numpy as np
import os
import random

def generate_premium_data(n_samples=5000, seed=42):
    """
    Generates synthetic premium training data calibrated with real-world Indian metropolitan baselines.
    
    Cities/Zones: Chennai, Bengaluru, Delhi
    Seasons: Monsoon, Summer, Winter, Normal
    """
    np.random.seed(seed)
    random.seed(seed)
    
    zones = ['Chennai_Urban', 'Bengaluru_Urban', 'Delhi_North', 'Delhi_South']
    vehicle_types = ['two_wheeler', 'three_wheeler', 'car']
    seasons = ['Monsoon', 'Summer', 'Winter', 'Normal']
    
    data = []
    
    for i in range(n_samples):
        worker_id = 1000 + i
        zone = random.choice(zones)
        vehicle = random.choice(vehicle_types)
        season = random.choice(seasons)
        
        # Base Risk Factors
        recent_disruption_rate = np.random.beta(2, 5) # Skewed towards lower disruption
        order_volatility = np.random.beta(2, 4)
        
        # City-Specific Meteorological Baselines
        rainfall_mm = 0.0
        temp_c = 28.0
        aqi = 50.0
        
        if zone == 'Chennai_Urban':
            temp_c = np.random.uniform(25, 35)
            if season == 'Monsoon': # Northeast Monsoon (Oct-Dec)
                rainfall_mm = np.random.exponential(40) # Occasional heavy bursts
            else:
                rainfall_mm = np.random.exponential(5)
            aqi = np.random.uniform(40, 120)
            
        elif zone == 'Bengaluru_Urban':
            temp_c = np.random.uniform(20, 30)
            if season == 'Monsoon':
                rainfall_mm = np.random.gamma(5, 5) # Persistent moderate rain
            else:
                rainfall_mm = np.random.exponential(8)
            aqi = np.random.uniform(50, 150)
            
        elif 'Delhi' in zone:
            if season == 'Summer':
                temp_c = np.random.uniform(35, 48) # Extreme Heat
                rainfall_mm = np.random.exponential(2)
            elif season == 'Winter':
                temp_c = np.random.uniform(5, 20)
                aqi = np.random.uniform(250, 500) # Severe Winter Pollution
            elif season == 'Monsoon':
                temp_c = np.random.uniform(28, 38)
                rainfall_mm = np.random.exponential(30)
            else:
                temp_c = np.random.uniform(22, 32)
                aqi = np.random.uniform(150, 250)

        # Logic for target 'premium_inr' and 'risk_score'
        # Formula: Base + (Rainfall Impact) + (Heat Impact) + (Pollution Impact) + (Volatility Impact)
        
        base_premium = 10.0
        risk_score = 0.1
        
        # Rainfall Impact (Heavy Rain > Payouts)
        rain_risk = min(1.0, rainfall_mm / 100.0)
        risk_score += rain_risk * 0.4
        
        # Heat Impact (Safety risk > disruption)
        heat_risk = 0.0
        if temp_c > 42:
            heat_risk = min(1.0, (temp_c - 42) / 10.0)
        risk_score += heat_risk * 0.2
        
        # Pollution Impact
        poll_risk = 0.0
        if aqi > 300:
            poll_risk = min(1.0, (aqi - 300) / 300.0)
        risk_score += poll_risk * 0.15
        
        # Volatility Impact
        risk_score += order_volatility * 0.2 + recent_disruption_rate * 0.3
        
        # Final Risk Score Bounded
        risk_score = min(1.0, risk_score)
        
        # Weekly Premium Calculation (Illustrative scale Rs 10 - Rs 50)
        premium_inr = base_premium + (risk_score * 40.0)
        
        # Vehicle Type Multiplier (higher cargo capacity usually = higher earnings = higher loss protection)
        if vehicle == 'three_wheeler':
            premium_inr *= 1.2
        elif vehicle == 'car':
            premium_inr *= 1.5
            
        data.append({
            'worker_id': worker_id,
            'zone_id': zone,
            'vehicle_type': vehicle,
            'season': season,
            'recent_disruption_rate': round(recent_disruption_rate, 3),
            'order_volatility': round(order_volatility, 3),
            'rainfall_mm': round(rainfall_mm, 2),
            'temp_c': round(temp_c, 1),
            'aqi': round(aqi, 0),
            'risk_score': round(risk_score, 3),
            'premium_inr': round(premium_inr, 2)
        })

    df = pd.DataFrame(data)
    
    # Save to ml/premium/data/
    output_dir = 'c:/Users/ritha/OneDrive/Documents/Amrita/Devtrails/InDel/ml/premium/data'
    os.makedirs(output_dir, exist_ok=True)
    
    output_path = os.path.join(output_dir, 'premium_training_data.csv')
    df.to_csv(output_path, index=False)
    
    print(f"Generated {n_samples} samples at {output_path}")
    print(df.head())

if __name__ == "__main__":
    generate_premium_data()
