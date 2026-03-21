from sklearn.cluster import DBSCAN
import numpy as np

class DBSCANDetector:
    def __init__(self):
        self.model = DBSCAN(eps=0.5, min_samples=5)
    
    def train(self, X):
        self.model.fit(X)
    
    def detect(self, instance):
        # Return cluster label
        return self.model.labels_.tolist()
