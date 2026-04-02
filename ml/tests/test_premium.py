import pytest
import pandas as pd
import os
import sys

# Ensure local imports work
sys.path.append(os.path.join(os.path.dirname(__file__), '../premium'))

from model import PremiumModel
from shap_explainer import SHAPExplainer

def test_model_loading_and_inference():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    model_path = os.path.join(script_dir, '../premium/artifacts/premium_model.joblib')
    assert os.path.exists(model_path), "Model artifact must exist. Run train.py first."
    
    model = PremiumModel.load(model_path)
    assert model is not None
    
    # Create a test sample
    test_df = pd.DataFrame([{
        'worker_id': 'wkr_test',
        'zone_id': 'zone_chennai_coastal',
        'city': 'Chennai',
        'state': 'Tamil Nadu',
        'zone_type': 'coastal',
        'vehicle_type': 'two_wheeler',
        'season': 'Monsoon',
        'experience_days': 500,
        'avg_daily_orders': 20.0,
        'avg_daily_earnings': 1200.0,
        'active_hours_per_day': 8.5,
        'rainfall_mm': 50.0,
        'aqi': 60.0,
        'temperature': 28.0,
        'humidity': 85.0,
        'order_volatility': 0.15,
        'earnings_volatility': 0.18,
        'recent_disruption_rate': 0.05
    }])
    
    premium, risk = model.predict(test_df)
    assert len(premium) == 1
    assert len(risk) == 1
    assert premium[0] > 0
    assert 0 <= risk[0] <= 1

def test_shap_explanation():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    model_path = os.path.join(script_dir, '../premium/artifacts/premium_model.joblib')
    model = PremiumModel.load(model_path)
    explainer = SHAPExplainer(model)
    
    test_df = pd.DataFrame([{
        'worker_id': 'wkr_test',
        'zone_id': 'zone_chennai_coastal',
        'city': 'Chennai',
        'state': 'Tamil Nadu',
        'zone_type': 'coastal',
        'vehicle_type': 'two_wheeler',
        'season': 'Monsoon',
        'experience_days': 500,
        'avg_daily_orders': 20.0,
        'avg_daily_earnings': 1200.0,
        'active_hours_per_day': 8.5,
        'rainfall_mm': 50.0,
        'aqi': 60.0,
        'temperature': 28.0,
        'humidity': 85.0,
        'order_volatility': 0.15,
        'earnings_volatility': 0.18,
        'recent_disruption_rate': 0.05
    }])
    
    explanation = explainer.explain(test_df)[0]
    
    # Check if all features are present in explainability
    features_in_explanation = [item['feature'] for item in explanation]
    for feature in model.feature_cols:
        assert feature in features_in_explanation
    
    # Check impact values are floats
    for item in explanation:
        assert isinstance(item['impact'], float)

if __name__ == "__main__":
    # If run directly, just run the tests
    test_model_loading_and_inference()
    test_shap_explanation()
    print("All unit tests passed!")
