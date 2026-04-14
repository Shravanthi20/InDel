import pandas as pd
import joblib
import os
from isolation_forest import IsolationForestDetector
from dbscan import DBSCANDetector

def train_models():
    """Train fraud detection models"""
    print("Loading synthetic dataset...")
    df = pd.read_csv('data/synthetic_claim_patterns.csv')
    
    # Select features for our unsupervised models
    features = [
        'earnings_drop_ratio', 
        'avg_orders_per_hour', 
        'distance_from_zone_center', 
        'claim_frequency', 
        'approval_ratio', 
        'zone_risk_score'
    ]
    
    X = df[features].values
    
    # Train Isolation Forest
    print("Training Isolation Forest...")
    iso_forest = IsolationForestDetector()
    iso_forest.train(X)
    
    # Train DBSCAN
    print("Training DBSCAN...")
    dbscan = DBSCANDetector()
    dbscan.train(X)
    
    # Make sure output directory exists
    os.makedirs('models', exist_ok=True)
    
    # Save the models using joblib
    print("Saving models...")
    joblib.dump(iso_forest.model, 'models/isolation_forest.joblib')
    joblib.dump(dbscan.model, 'models/dbscan.joblib')
    
    print("Fraud detection models training and saving complete!")

if __name__ == "__main__":
    train_models()
