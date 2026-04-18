package main

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

func main() {
	brokers := "d7h2gvf095r0u8rsa8m0.any.us-east-1.mpx.prd.cloud.redpanda.com:9092"
	user := "indel-the-king"
	pass := "PAw1BJMw22FtKcosL3p5Wr7Rogd6yC"
	topic := "claim.disruption.created"

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.SASL(scram.Auth{User: user, Pass: pass}.AsSha512Mechanism()),
		kgo.DialTLS(),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte("test-key"),
		Value: []byte("hello from franz-go!"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
		panic(err)
	}
	fmt.Println("Message sent successfully!")
}
