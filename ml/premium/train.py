import pandas as pd
from model import PremiumModel
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_absolute_error
import os

def train_model():
    """Train premium pricing and risk models"""
    script_dir = os.path.dirname(os.path.abspath(__file__))
    data_path = os.path.join(script_dir, 'data/premium_training_data_india.csv')
    if not os.path.exists(data_path):
        print(f"Error: Data file not found at {data_path}")
        return

    # Load synthetic data
    df = pd.read_csv(data_path)
    
    # Prepare features and targets
    # Drop non-feature columns
    X = df.drop(columns=['worker_id', 'premium_inr', 'risk_score'])
    y_premium = df['premium_inr']
    y_risk = df['risk_score']
    
    # Simple split
    X_train, X_test, y_p_train, y_p_test, y_r_train, y_r_test = train_test_split(
        X, y_premium, y_risk, test_size=0.2, random_state=42
    )
    
    model = PremiumModel()
    model.train(X_train, y_p_train, y_r_train)
    
    # Evaluate
    p_pred, r_pred = model.predict(X_test)
    mae_premium = mean_absolute_error(y_p_test, p_pred)
    mae_risk = mean_absolute_error(y_r_test, r_pred)
    
    print(f"Premium MAE: {mae_premium:.4f}")
    print(f"Risk Multiplier MAE: {mae_risk:.4f}")
    
    # Save artifacts
    artifacts_path = os.path.join(script_dir, 'artifacts/premium_model.joblib')
    model.save(artifacts_path)
    print(f"Training complete. Artifacts saved to {artifacts_path}")

if __name__ == "__main__":
    train_model()
