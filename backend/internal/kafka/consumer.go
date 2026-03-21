package kafka

import (
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
