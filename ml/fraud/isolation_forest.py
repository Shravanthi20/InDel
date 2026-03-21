from sklearn.ensemble import IsolationForest
import numpy as np

class IsolationForestDetector:
    def __init__(self):
        self.model = IsolationForest(contamination=0.05)
    
    def train(self, X):
        self.model.fit(X)
    
    def detect(self, instance):
        score = self.model.decision_function(instance.reshape(1, -1))[0]
        return score
