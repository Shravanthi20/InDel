package workers

// fraud_consumer.go - Consumes fraud scores
type FraudConsumer struct{}

func (f *FraudConsumer) Start() error {
	// Subscribe to indel.claims.scored
	// Update claim fraud status
	return nil
}
