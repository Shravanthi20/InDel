package pollers

// aqi_poller.go - Polls OpenAQ API every 30 minutes per zone
type AQIPoller struct{}

func (p *AQIPoller) Start() error {
	// Poll OpenAQ for each zone every 30 minutes
	// Publish to indel.aqi.alerts if AQI high
	return nil
}
