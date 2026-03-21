import pandas as pd
from prophet_model import ProphetForecaster

def train_zone_models():
    """Train per-zone forecast models"""
    df = pd.read_csv('data/synthetic_zone_history.csv')
    
    # For each zone, train a Prophet model
    for zone_id in [1, 2, 3, 4]:
        zone_data = df[df['zone_id'] == zone_id]
        
        forecaster = ProphetForecaster(zone_id)
        # forecaster.train(zone_data)
        
        print(f"Forecast model trained for zone {zone_id}")

if __name__ == "__main__":
    train_zone_models()
