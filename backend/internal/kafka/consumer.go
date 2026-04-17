package kafka

import (
	"os"
	"strings"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
}

func NewConsumer(brokers string, group string, topics []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Return.Errors = true

	// SASL/PLAIN authentication for Redpanda/Confluent Cloud
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")
	if kafkaUser != "" && kafkaPass != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaUser
		config.Net.SASL.Password = kafkaPass
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.TLS.Enable = true // Redpanda Cloud requires TLS
	}

	brokerList := strings.Split(brokers, ",")
	consumerGroup, err := sarama.NewConsumerGroup(brokerList, group, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: consumerGroup}, nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
