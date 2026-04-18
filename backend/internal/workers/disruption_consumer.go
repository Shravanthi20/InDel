package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/Shravanthi20/InDel/backend/internal/events"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
)

// DisruptionConsumer consumes disruption events from Kafka
type DisruptionConsumer struct {
	Brokers string
	Group   string
}

func (d *DisruptionConsumer) Start() error {
	topics := []string{kafka.TopicClaimDisruptionCreated, kafka.TopicClaimDisruptionUpdated}
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup([]string{d.Brokers}, d.Group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &disruptionEventHandler{}

	// Handle SIGINT/SIGTERM for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	log.Printf("[KAFKA] DisruptionConsumer started. Listening on topics: %v", topics)
	for {
		if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
			log.Printf("[KAFKA] Consumer error: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil
}

type disruptionEventHandler struct{}

func (h *disruptionEventHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *disruptionEventHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *disruptionEventHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var evt events.ClaimDisruptionEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("[KAFKA] Failed to unmarshal disruption event: %v", err)
			continue
		}
		log.Printf("[KAFKA] Disruption event consumed: %+v", evt)

		// --- Recovery Logic ---
		// 1. Simulate DB update for disruption state
		err := updateDisruptionState(evt)
		if err != nil {
			log.Printf("[RECOVERY] DB update failed for claim %s: %v", evt.ClaimID, err)
			// 2. Retry logic (simple exponential backoff, max 3 attempts)
			for attempt := 1; attempt <= 3; attempt++ {
				log.Printf("[RECOVERY] Retrying DB update for claim %s (attempt %d)", evt.ClaimID, attempt)
				time.Sleep(time.Duration(attempt*200) * time.Millisecond)
				if err := updateDisruptionState(evt); err == nil {
					log.Printf("[RECOVERY] DB update succeeded for claim %s on retry %d", evt.ClaimID, attempt)
					break
				} else if attempt == 3 {
					log.Printf("[RECOVERY] All retries failed for claim %s. Triggering fallback.", evt.ClaimID)
					triggerFallback(evt)
				}
			}
		} else {
			log.Printf("[RECOVERY] DB update succeeded for claim %s", evt.ClaimID)
		}

		// 3. Alerting (simulate alert if severity is high)
		if evt.Severity == "high" {
			triggerAlert(evt)
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

// --- Simulated recovery helpers ---
func updateDisruptionState(evt events.ClaimDisruptionEvent) error {
	// Simulate DB update (random failure for demo)
	if evt.Metadata != nil && evt.Metadata["simulate_db_fail"] == true {
		return fmt.Errorf("simulated DB failure")
	}
	// In real code: update disruption/claim state in DB
	return nil
}

func triggerFallback(evt events.ClaimDisruptionEvent) {
	log.Printf("[FALLBACK] Fallback triggered for claim %s (type: %s)", evt.ClaimID, evt.DisruptionType)
	// In real code: enqueue for manual review, write to fallback queue, etc.
}

func triggerAlert(evt events.ClaimDisruptionEvent) {
	log.Printf("[ALERT] High severity disruption for claim %s: type=%s, metadata=%v", evt.ClaimID, evt.DisruptionType, evt.Metadata)
	// In real code: send alert to ops/oncall, push notification, etc.
}
