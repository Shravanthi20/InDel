package pollers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/services"
)

// PayoutConsumer listens for payout.queued events on Kafka and processes them.
type PayoutConsumer struct {
	Consumer *kafka.Consumer
	CoreSvc  *services.CoreOpsService
}

func (p *PayoutConsumer) Start() error {
	go func() {
		ctx := context.Background()
		for {
			err := p.Consumer.Subscribe(ctx, []string{kafka.TopicPayoutsQueued}, p)
			if err != nil {
				log.Printf("[PayoutConsumer] Error: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (p *PayoutConsumer) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (p *PayoutConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (p *PayoutConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event struct {
			Payload struct {
				PayoutID uint `json:"payout_id"`
			} `json:"payload"`
		}

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[PayoutConsumer] JSON error: %v", err)
			continue
		}

		if event.Payload.PayoutID == 0 {
			log.Printf("[PayoutConsumer] Invalid payout_id in message")
			continue
		}

		log.Printf("[PayoutConsumer] Processing payout_%d", event.Payload.PayoutID)
		_, err := p.CoreSvc.ProcessQueuedPayouts(time.Now().UTC()) // This processes all queued including retry_pending
		// More targeted approach would be s.processPayoutsByID([]uint{event.Payload.PayoutID}, now)
		// but ProcessQueuedPayouts is already public and safe.
		
		if err != nil {
			log.Printf("[PayoutConsumer] Failed to process payout_%d: %v", event.Payload.PayoutID, err)
		} else {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
