import pandas as pd
import numpy as np
import os
import random
import uuid

def generate_india_premium_data(n_samples=30000, seed=42):
    """
    Generates a large-scale synthetic training dataset for the InDel Premium Prediction model,
    covering 12 Indian cities across 5 diverse zone types.
    
    Premium Scaling: ₹20 - ₹50
    """
    np.random.seed(seed)
    random.seed(seed)
    
    # Configuration: City to Zone and State Mapping
    city_config = {
        'Chennai': {'type': 'coastal', 'state': 'Tamil Nadu'},
        'Mumbai': {'type': 'coastal', 'state': 'Maharashtra'},
        'Kolkata': {'type': 'coastal', 'state': 'West Bengal'},
        'Delhi': {'type': 'pollution_heavy', 'state': 'Delhi'},
        'Bengaluru': {'type': 'urban', 'state': 'Karnataka'},
        'Hyderabad': {'type': 'urban', 'state': 'Telangana'},
        'Pune': {'type': 'urban', 'state': 'Maharashtra'},
        'Ahmedabad': {'type': 'dry', 'state': 'Gujarat'},
        'Jaipur': {'type': 'dry', 'state': 'Rajasthan'},
        'Coimbatore': {'type': 'tier2', 'state': 'Tamil Nadu'},
        'Lucknow': {'type': 'tier2', 'state': 'Uttar Pradesh'},
        'Indore': {'type': 'tier2', 'state': 'Madhya Pradesh'}
    }
    
    cities = list(city_config.keys())
    vehicle_types = ['two_wheeler', 'bike', 'car']
    seasons = ['Summer', 'Monsoon', 'Winter']
    
    data = []
    
    for i in range(n_samples):
        # 1. Geographic & Temporal Features
        city = random.choice(cities)
        config = city_config[city]
        zone_type = config['type']
        state = config['state']
        zone_id = f"zone_{city.lower()}_{zone_type}"
        season = random.choice(seasons)
        
        # 2. Worker Features
        worker_id = f"wkr_{10000 + i}"
        vehicle = random.choice(vehicle_types)
        experience_days = random.randint(30, 1500)
        
        # Base activity levels
        active_hours = np.random.normal(8, 2)
        active_hours = max(4, min(14, active_hours))
        
        avg_daily_orders = np.random.normal(15, 5)
        avg_daily_orders = max(5, min(40, avg_daily_orders))
        
        # Earnings correlated with orders and vehicle
        avg_daily_earnings = avg_daily_orders * np.random.uniform(30, 60)
        if vehicle == 'car':
            avg_daily_earnings *= 1.4
        elif vehicle == 'two_wheeler':
            avg_daily_earnings *= 1.1
            
        # 3. Environmental Features (Rule-Based)
        rainfall_mm = 0.0
        aqi = 50.0
        temp_c = 28.0
        humidity = 60.0
        
        # Seasonal Baseline
        if season == 'Summer':
            temp_c = np.random.uniform(32, 42)
            humidity = np.random.uniform(20, 50)
            rainfall_mm = np.random.exponential(2)
        elif season == 'Monsoon':
            temp_c = np.random.uniform(25, 32)
            humidity = np.random.uniform(70, 95)
            rainfall_mm = np.random.exponential(15)
        else: # Winter
            temp_c = np.random.uniform(15, 25)
            humidity = np.random.uniform(40, 60)
            rainfall_mm = np.random.exponential(1)

        # Zone-Specific Modifiers
        if zone_type == 'coastal':
            humidity += np.random.uniform(10, 20)
            if season == 'Monsoon':
                rainfall_mm = np.random.gamma(5, 15) # High rainfall bursts (30-120mm range)
            aqi = np.random.uniform(30, 100)
            
        elif zone_type == 'pollution_heavy':
            if season == 'Winter':
                aqi = np.random.uniform(250, 500) # Severe Delhi Winters
            else:
                aqi = np.random.uniform(150, 300)
            temp_c += np.random.uniform(2, 5) if season == 'Summer' else 0
            
        elif zone_type == 'dry':
            humidity -= np.random.uniform(15, 30)
            temp_c += np.random.uniform(3, 8) if season == 'Summer' else 0
            rainfall_mm *= 0.2
            aqi = np.random.uniform(60, 180)
            
        elif zone_type == 'urban':
            aqi = np.random.uniform(80, 220)
            avg_daily_orders += 5 # Higher demand
            
        elif zone_type == 'tier2':
            aqi = np.random.uniform(50, 150)
            avg_daily_orders -= 3
            
        # Bounds check for environment
        rainfall_mm = max(0.0, min(250.0, rainfall_mm))
        aqi = max(10.0, min(500.0, aqi))
        humidity = max(10.0, min(100.0, humidity))

        # 4. Behavioral Features (Correlation Logic)
        # Disruption increases with high Rain or high AQI
        env_stress = (rainfall_mm / 100.0) * 0.6 + (aqi / 500.0) * 0.4
        recent_disruption_rate = np.random.beta(2 + 5 * env_stress, 8)
        
        # Volatility increases with disruption and zone risk
        order_volatility = np.random.beta(2 + 3 * env_stress, 6)
        earnings_volatility = order_volatility * np.random.uniform(0.8, 1.2)
        
        # 5. Risk Score Calculation (Weighted)
        # Weights: Rain (0.35), AQI (0.25), Disruption (0.25), Volatility (0.15)
        rain_norm = min(1.0, rainfall_mm / 100.0)
        aqi_norm = min(1.0, aqi / 500.0)
        
        risk_score = (rain_norm * 0.35) + \
                     (aqi_norm * 0.25) + \
                     (recent_disruption_rate * 0.25) + \
                     (order_volatility * 0.15)
        
        risk_score = min(1.0, max(0.0, risk_score + np.random.normal(0, 0.02)))
        
        # 6. Premium Calculation (₹20 - ₹50)
        # Base: 20, Range: 30
        premium_inr = 20.0 + (risk_score * 30.0)
        
        data.append({
            'worker_id': worker_id,
            'vehicle_type': vehicle,
            'experience_days': experience_days,
            'avg_daily_orders': round(avg_daily_orders, 1),
            'avg_daily_earnings': round(avg_daily_earnings, 2),
            'active_hours_per_day': round(active_hours, 1),
            'zone_id': zone_id,
            'city': city,
            'state': state,
            'zone_type': zone_type,
            'season': season,
            'rainfall_mm': round(rainfall_mm, 2),
            'aqi': round(aqi, 0),
            'temperature': round(temp_c, 1),
            'humidity': round(humidity, 1),
            'order_volatility': round(order_volatility, 3),
            'earnings_volatility': round(earnings_volatility, 3),
            'recent_disruption_rate': round(recent_disruption_rate, 3),
            'risk_score': round(risk_score, 4),
            'premium_inr': round(premium_inr, 2)
        })

    df = pd.DataFrame(data)
    
    # Save to CSV
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_path = os.path.join(script_dir, '../premium/data/premium_training_data_india.csv')
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    df.to_csv(output_path, index=False)
    
    print(f"Generated {n_samples} samples at {output_path}")
    
    # Brief explanation logic
    print("\nFeature Generation Logic Summary:")
    print("1. Rainfall: Coastal zones in Monsoon follow a Gamma distribution (high intensity).")
    print("2. AQI: Pollution-heavy zones (Delhi) in Winter peak between 250-500.")
    print("3. Behavior: Disruption and Volatility are dynamically correlated with Environmental Stress (Rain/AQI).")
    print("4. Risk Score: Weighted mapping of normalized Rain, AQI, and Behavioral signals.")
    print("5. Premium: Scaled to INR 20 - INR 50 based on the calculated risk score.")

if __name__ == "__main__":
    generate_india_premium_data()
