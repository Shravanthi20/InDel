import pandas as pd

def engineer_features(raw_data):
    """Feature engineering for forecasting"""
    # Extract weather, AQI, order volume features
    features = {}
    
    features['avg_temperature'] = raw_data.get('temperature', 25)
    features['aqi'] = raw_data.get('aqi', 50)
    features['order_volume'] = raw_data.get('order_volume', 1000)
    
    return features
