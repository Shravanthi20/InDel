package kafka

import (
	"os"
	"testing"
)

func TestKafkaBrokerConnection(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("KAFKA_BROKERS env not set")
	}

	producer, err := NewProducer(brokers)
	if err != nil {
		t.Fatalf("Failed to connect to Kafka brokers %s: %v", brokers, err)
	}
	defer producer.Close()
}
