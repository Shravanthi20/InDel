import shap
import numpy as np

class SHAPExplainer:
    def __init__(self, model):
        self.explainer = shap.TreeExplainer(model)
    
    def explain(self, instance):
        shap_values = self.explainer.shap_values(instance)
        return {
            "shap_values": shap_values.tolist(),
            "base_value": self.explainer.expected_value,
            "explanation": self._get_plain_language_explanation(shap_values)
        }
    
    def _get_plain_language_explanation(self, shap_values):
        return "Premium calculation explanation based on SHAP values"
