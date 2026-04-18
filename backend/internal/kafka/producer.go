package kafka

import (
	"crypto/sha256"
	"hash"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/xdg/scram"
)

type Producer struct {
	producer sarama.AsyncProducer
}

func NewProducer(brokers string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0

	// SCRAM-SHA-256 authentication for Redpanda Cloud
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")
	if kafkaUser != "" && kafkaPass != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaUser
		config.Net.SASL.Password = kafkaPass
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		config.Net.SASL.Handshake = true
		config.Net.TLS.Enable = true
		config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
		}
	}

	brokerList := strings.Split(brokers, ",")
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

// SCRAM client implementation for SHA-256
var SHA256 scram.HashGeneratorFcn = func() hash.Hash { return sha256.New() }

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	HashGeneratorFcn scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
       var err error
       x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
       if err != nil {
	       return err
       }
       x.ClientConversation = x.Client.NewConversation()
       return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
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
