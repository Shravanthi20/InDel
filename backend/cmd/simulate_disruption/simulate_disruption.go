package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/Shravanthi20/InDel/backend/internal/events"
)

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		fmt.Println("KAFKA_BROKERS must be set in env")
		os.Exit(1)
	}
	topic := "claim.disruption.created"

	event := events.ClaimDisruptionEvent{
		EventType:      "claim.disruption.created",
		ClaimID:        fmt.Sprintf("sim-%d", time.Now().Unix()),
		DisruptionType: "heavy_rain",
		Timestamp:      time.Now().UTC(),
		Severity:       "high",
		Metadata: map[string]interface{}{
			"simulate_db_fail": false,
			"note":             "Simulated from Go script",
		},
	}

	b, _ := json.Marshal(event)

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	// SASL/PLAIN authentication for cloud Kafka
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")
	if kafkaUser != "" && kafkaPass != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaUser
		config.Net.SASL.Password = kafkaPass
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ClaimID),
		Value: sarama.ByteEncoder(b),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	fmt.Println("Simulated disruption event sent to Kafka!")
}
