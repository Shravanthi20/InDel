package kafka

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/Shravanthi20/InDel/backend/internal/events"
)

func TestDisruptionEventEmissionAndConsumption(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("KAFKA_BROKERS env not set")
	}
	topic := "claim.disruption.created"

	group := "test-disruption-consumer-group"

	event := events.ClaimDisruptionEvent{
		EventType:      "claim.disruption.created",
		ClaimID:        "test-claim-123",
		DisruptionType: "test_disruption",
		Timestamp:      time.Now().UTC(),
		Severity:       "high",
		Metadata:       map[string]interface{}{"test": true},
	}
	b, _ := json.Marshal(event)

	// Start consumer
	consumerReady := make(chan struct{})
	received := make(chan events.ClaimDisruptionEvent, 1)
	go func() {
		config := sarama.NewConfig()
		config.Version = sarama.V2_6_0_0
		config.Consumer.Return.Errors = true
		cg, err := sarama.NewConsumerGroup([]string{brokers}, group, config)
		if err != nil {
			t.Errorf("Failed to create consumer group: %v", err)
			return
		}
		defer cg.Close()
		handler := &testDisruptionHandler{received: received, ready: consumerReady}
		go func() {
			_ = cg.Consume(context.Background(), []string{topic}, handler)
		}()
		<-consumerReady // Wait for consumer to be ready
		// Wait for message or timeout
		select {
		case <-time.After(10 * time.Second):
			t.Error("Timed out waiting for disruption event")
		case <-received:
			// Success
		}
	}()

	<-consumerReady // Wait for consumer to be ready

	// Emit event
	producer, err := NewProducer(brokers)
	if err != nil {
		t.Fatalf("Failed to connect to Kafka brokers %s: %v", brokers, err)
	}
	defer producer.Close()
	if err := producer.Publish(topic, event.ClaimID, b); err != nil {
		t.Fatalf("Failed to publish disruption event: %v", err)
	}

	// Wait for consumer to receive
	time.Sleep(3 * time.Second)
}

type testDisruptionHandler struct {
	received chan events.ClaimDisruptionEvent
	ready    chan struct{}
}

func (h *testDisruptionHandler) Setup(_ sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}
func (h *testDisruptionHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *testDisruptionHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var evt events.ClaimDisruptionEvent
		_ = json.Unmarshal(msg.Value, &evt)
		h.received <- evt
		sess.MarkMessage(msg, "")
	}
	return nil
}
