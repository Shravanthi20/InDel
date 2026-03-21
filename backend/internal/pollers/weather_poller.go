package pollers

// weather_poller.go - Polls OpenWeatherMap API every 10 minutes per zone
type WeatherPoller struct{}

func (p *WeatherPoller) Start() error {
	// Poll OpenWeatherMap for each zone every 10 minutes
	// Publish to indel.weather.alerts if disruption detected
	return nil
}
