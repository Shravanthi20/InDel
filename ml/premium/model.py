import numpy as np
import pandas as pd
from xgboost import XGBRegressor
from sklearn.preprocessing import LabelEncoder
import joblib
import os

class PremiumModel:
    def __init__(self):
        self.premium_model = XGBRegressor(
            n_estimators=100,
            learning_rate=0.1,
            max_depth=5,
            random_state=42
        )
        self.risk_model = XGBRegressor(
            n_estimators=100,
            learning_rate=0.1,
            max_depth=5,
            random_state=42
        )
        self.label_encoders = {}
        self.categorical_cols = ['zone_id', 'city', 'state', 'zone_type', 'vehicle_type', 'season']
        self.feature_cols = [
            'zone_id', 'city', 'state', 'zone_type', 'vehicle_type', 'season',
            'experience_days', 'avg_daily_orders', 'avg_daily_earnings', 'active_hours_per_day',
            'rainfall_mm', 'aqi', 'temperature', 'humidity',
            'order_volatility', 'earnings_volatility', 'recent_disruption_rate'
        ]

    def _preprocess(self, df, training=False):
        """Encodes categorical variables."""
        df_proc = df.copy()
        for col in self.categorical_cols:
            if training:
                le = LabelEncoder()
                df_proc[col] = le.fit_transform(df_proc[col])
                self.label_encoders[col] = le
            else:
                le = self.label_encoders[col]
                # Handle unseen categorical values at inference by mapping
                # them to the first known class for that feature.
                known_classes = set(le.classes_)
                fallback = le.classes_[0]
                df_proc[col] = df_proc[col].apply(
                    lambda v: v if v in known_classes else fallback
                )
                df_proc[col] = le.transform(df_proc[col])
        return df_proc[self.feature_cols]

    def train(self, X_df, y_premium, y_risk):
        """X_df: DataFrame with raw features, y_premium: premium target, y_risk: risk target."""
        X_processed = self._preprocess(X_df, training=True)
        self.premium_model.fit(X_processed, y_premium)
        self.risk_model.fit(X_processed, y_risk)
        print("Models trained successfully.")

    def predict(self, X_df):
        X_processed = self._preprocess(X_df, training=False)
        premium = self.premium_model.predict(X_processed)
        risk = self.risk_model.predict(X_processed)
        return premium, risk

    def save(self, path):
        os.makedirs(os.path.dirname(path), exist_ok=True)
        joblib.dump({
            'premium_model': self.premium_model,
            'risk_model': self.risk_model,
            'label_encoders': self.label_encoders,
            'categorical_cols': self.categorical_cols,
            'feature_cols': self.feature_cols
        }, path)
        print(f"Model artifacts saved to {path}")

    @classmethod
    def load(cls, path):
        artifacts = joblib.load(path)
        instance = cls()
        instance.premium_model = artifacts['premium_model']
        instance.risk_model = artifacts['risk_model']
        instance.label_encoders = artifacts['label_encoders']
        instance.categorical_cols = artifacts['categorical_cols']
        instance.feature_cols = artifacts['feature_cols']
        return instance
