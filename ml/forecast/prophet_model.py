from fbprophet import Prophet
import pandas as pd

class ProphetForecaster:
    def __init__(self, zone_id):
        self.zone_id = zone_id
        self.model = None
    
    def train(self, historical_data):
        # historical_data: df with 'ds' (date) and 'y' (disruption_events)
        self.model = Prophet()
        # self.model.fit(historical_data)
    
    def forecast(self, periods=7):
        # future = self.model.make_future_dataframe(periods=periods)
        # forecast = self.model.predict(future)
        return []
