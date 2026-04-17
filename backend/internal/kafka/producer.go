package kafka

import (
	"os"
	"strings"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.AsyncProducer
}

func NewProducer(brokers string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0

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
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) Publish(topic string, key string, message []byte) error {
	p.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
	}
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
