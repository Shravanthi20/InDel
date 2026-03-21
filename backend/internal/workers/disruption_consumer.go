package workers

// disruption_consumer.go - Consumes disruption confirmations
type DisruptionConsumer struct{}

func (d *DisruptionConsumer) Start() error {
	// Subscribe to indel.disruption.confirmed
	// Generate auto claims
	return nil
}
