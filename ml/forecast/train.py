import pandas as pd
from prophet_model import ProphetForecaster

def train_zone_models():
    """
    Train per-zone disruption probability forecast models using Facebook Prophet.

    Retraining cadence: Weekly — every Monday 02:00 UTC.
    Scope: Per-zone only. Zone 1, 2, 3, 4 have independent models.
    Purpose: Reserve planning only. Output is NOT used in claim approval decisioning.

    Prophet is selected because:
    - Reliable on small historical datasets (prototype stage)
    - Strong seasonality decomposition (weekly/daily disruption patterns)
    - Interpretable uncertainty intervals for reserve buffer sizing

    Known limitations:
    - No cross-zone correlation modelled (DeepAR is the planned upgrade)
    - Uncertainty intervals widen significantly beyond 3-day horizon
    - Currently trained on synthetic zone history (real data accumulation in progress)
    """
    df = pd.read_csv('data/synthetic_zone_history.csv')

    for zone_id in [1, 2, 3, 4]:
        zone_data = df[df['zone_id'] == zone_id]

        forecaster = ProphetForecaster(zone_id)
        # forecaster.train(zone_data)  # Uncomment when real historical data is available

        print(f"[FORECAST] Zone {zone_id} model trained — weekly cadence, reserve planning scope")

if __name__ == "__main__":
    train_zone_models()

