import numpy as np

def engineer_features(raw_data):
    """Feature engineering for premium prediction"""
    features = {}
    
    # Calculate earnings volatility
    features['earnings_volatility'] = np.std(raw_data.get('earnings', []))
    
    # Calculate disruption frequency
    features['disruption_frequency'] = len(raw_data.get('disruptions', []))
    
    # Zone risk rating
    features['zone_risk_rating'] = raw_data.get('zone_risk', 0.5)
    
    # Account age
    features['account_age_days'] = raw_data.get('account_age_days', 0)
    
    return features
