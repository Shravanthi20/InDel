package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// Minimal disruption event struct for demo
// (replace with your full struct if needed)
type ClaimDisruptionEvent struct {
	EventType      string                 `json:"event_type"`
	ClaimID        string                 `json:"claim_id"`
	DisruptionType string                 `json:"disruption_type"`
	Timestamp      time.Time              `json:"timestamp"`
	Severity       string                 `json:"severity"`
	Metadata       map[string]interface{} `json:"metadata"`
}

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		fmt.Println("KAFKA_BROKERS must be set in env")
		os.Exit(1)
	}
	user := os.Getenv("KAFKA_USER")
	pass := os.Getenv("KAFKA_PASS")
	if user == "" || pass == "" {
		fmt.Println("KAFKA_USER and KAFKA_PASS must be set in env")
		os.Exit(1)
	}
	topic := "claim.disruption.created"

	event := ClaimDisruptionEvent{
		EventType:      "claim.disruption.created",
		ClaimID:        fmt.Sprintf("sim-%d", time.Now().Unix()),
		DisruptionType: "heavy_rain",
		Timestamp:      time.Now().UTC(),
		Severity:       "high",
		Metadata: map[string]interface{}{
			"note": "Simulated from franz-go script",
		},
	}
	b, _ := json.Marshal(event)

	// Enable franz-go debug logging to stderr
	logger := kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, nil)
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.SASL(scram.Auth{User: user, Pass: pass}.AsSha256Mechanism()), // Use SCRAM-SHA-256
		kgo.DialTLS(),
		kgo.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(event.ClaimID),
		Value: b,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
		panic(err)
	}
	fmt.Println("Simulated disruption event sent to Redpanda Cloud Kafka!")
}
