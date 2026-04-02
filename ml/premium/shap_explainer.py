import shap
import pandas as pd
import numpy as np

class SHAPExplainer:
    def __init__(self, premium_model):
        """
        premium_model: An instance of PremiumModel with the trained XGBoost and label encoders.
        """
        self.premium_model = premium_model
        # Use TreeExplainer for XGBoost (premium model)
        self.explainer = shap.TreeExplainer(self.premium_model.premium_model)
    
    def explain(self, X_df):
        """
        X_df: Raw DataFrame (before preprocessing).
        Returns a list of dicts with feature impacts.
        """
        # Preprocess features using the model's internal method
        X_processed = self.premium_model._preprocess(X_df, training=False)
        shap_values = self.explainer.shap_values(X_processed)
        
        # In XGBoost regressor, shap_values is a simple array of same shape as X_processed
        # For a single prediction:
        results = []
        for i in range(len(X_df)):
            feature_impacts = []
            for col_idx, col_name in enumerate(self.premium_model.feature_cols):
                impact = float(shap_values[i][col_idx])
                feature_impacts.append({
                    "feature": col_name,
                    "impact": round(impact, 3)
                })
            results.append(feature_impacts)
            
        return results
