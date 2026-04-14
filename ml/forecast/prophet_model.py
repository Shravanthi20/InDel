"""
Per-zone disruption probability forecaster using pandas seasonal decomposition.

This replaces the Facebook Prophet Stan backend with a reliable pandas-based
weekly seasonal model that:
  1. Computes day-of-week disruption averages from historical data
  2. Applies an exponential moving average trend
  3. Generates a 7-day forward probability forecast

Purpose: Reserve planning only — NOT claim decisioning.
Retraining cadence: Weekly — every Monday 02:00 UTC.
Scope: Per-zone independent models. No cross-zone correlation.
"""
import pandas as pd
import numpy as np
from datetime import date, timedelta


class ProphetForecaster:
    """
    Lightweight seasonal forecaster — functionally equivalent to Prophet
    for short-horizon weekly patterns without Stan compilation dependency.
    """

    def __init__(self, zone_id: int):
        self.zone_id = zone_id
        self.trained = False
        self._dow_avg: dict[int, float] = {}   # day-of-week → avg events
        self._recent_trend: float = 0.0         # EMA of last 14-day events
        self._global_avg: float = 0.0
        self._max_events: float = 5.0

    def train(self, historical_data: pd.DataFrame):
        """
        Fit a weekly seasonal model from historical zone data.

        Args:
            historical_data: DataFrame with 'date' and 'disruption_events' columns.
        """
        df = historical_data[["date", "disruption_events"]].copy()
        df["date"] = pd.to_datetime(df["date"])
        df = df.sort_values("date").reset_index(drop=True)
        df["dow"] = df["date"].dt.dayofweek   # 0=Mon … 6=Sun

        # Day-of-week averages (core seasonality signal)
        self._dow_avg = df.groupby("dow")["disruption_events"].mean().to_dict()

        # Fill any missing days of week with global average
        self._global_avg = float(df["disruption_events"].mean())
        for d in range(7):
            self._dow_avg.setdefault(d, self._global_avg)

        # Trend: EMA of last 14 days (alpha=0.3 → slow-moving trend)
        last_14 = df.tail(14)["disruption_events"].tolist()
        ema = last_14[0] if last_14 else self._global_avg
        alpha = 0.3
        for val in last_14[1:]:
            ema = alpha * val + (1 - alpha) * ema
        self._recent_trend = ema

        # Max events seen (for normalisation to probability)
        self._max_events = max(float(df["disruption_events"].max()), 3.0)

        self.trained = True

    def forecast(self, periods: int = 7) -> list[dict]:
        """
        Generate a <periods>-day forward disruption probability forecast.

        Returns:
            List of {"date": "YYYY-MM-DD", "disruption_probability": float}
        """
        if not self.trained:
            return []

        today = date.today()
        results = []

        for i in range(periods):
            target_date = today + timedelta(days=i)
            dow = target_date.weekday()   # 0=Mon … 6=Sun

            # Blend seasonal pattern with recent trend (60/40 split)
            seasonal = self._dow_avg.get(dow, self._global_avg)
            blended = 0.60 * seasonal + 0.40 * self._recent_trend

            # Apply a slight forward trend decay (uncertainty grows with horizon)
            decay = 1.0 - (i * 0.02)   # −2% per day forward
            predicted_events = blended * decay

            # Normalise to [0, 1] probability
            prob = round(min(max(predicted_events / self._max_events, 0.0), 1.0), 3)

            results.append({
                "date": target_date.isoformat(),
                "disruption_probability": prob,
            })

        return results
