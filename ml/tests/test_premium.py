import pytest
import pandas as pd
import os
import sys

# Ensure local imports work
sys.path.append(os.path.join(os.path.dirname(__file__), '../premium'))

from model import PremiumModel
from shap_explainer import SHAPExplainer

def test_model_loading_and_inference():
    model_path = 'c:/Users/ritha/OneDrive/Documents/Amrita/Devtrails/InDel/ml/premium/artifacts/premium_model.joblib'
    assert os.path.exists(model_path), "Model artifact must exist. Run train.py first."
    
    model = PremiumModel.load(model_path)
    assert model is not None
    
    # Create a test sample
    test_df = pd.DataFrame([{
        'worker_id': 'wkr_test',
        'zone_id': 'Chennai_Urban',
        'vehicle_type': 'two_wheeler',
        'season': 'Monsoon',
        'recent_disruption_rate': 0.5,
        'order_volatility': 0.3,
        'rainfall_mm': 60.0,
        'temp_c': 28.0,
        'aqi': 80.0
    }])
    
    premium, risk = model.predict(test_df)
    assert len(premium) == 1
    assert len(risk) == 1
    assert premium[0] > 0
    assert 0 <= risk[0] <= 1

def test_shap_explanation():
    model_path = 'c:/Users/ritha/OneDrive/Documents/Amrita/Devtrails/InDel/ml/premium/artifacts/premium_model.joblib'
    model = PremiumModel.load(model_path)
    explainer = SHAPExplainer(model)
    
    test_df = pd.DataFrame([{
        'worker_id': 'wkr_test',
        'zone_id': 'Chennai_Urban',
        'vehicle_type': 'two_wheeler',
        'season': 'Monsoon',
        'recent_disruption_rate': 0.5,
        'order_volatility': 0.3,
        'rainfall_mm': 60.0,
        'temp_c': 28.0,
        'aqi': 80.0
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
