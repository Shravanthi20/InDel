package pollers

import (
	"log"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/gorm"
)

// DisruptionProcessor picks up newly confirmed disruptions and runs the
// full pipeline: notify workers → generate claims → queue payouts → process payouts.
// Runs every 2 minutes.
type DisruptionProcessor struct {
	DB      *gorm.DB
	CoreSvc *services.CoreOpsService
}

func (p *DisruptionProcessor) Start() {
	go func() {
		p.process()
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			p.process()
		}
	}()
}

func (p *DisruptionProcessor) process() {
	if p.DB == nil {
		return
	}

	// Find disruptions confirmed in the last 10 minutes that haven't been processed yet
	cutoff := time.Now().UTC().Add(-10 * time.Minute)
	var disruptions []models.Disruption
	err := p.DB.Where(
		"status = ? AND confirmed_at >= ? AND processed_at IS NULL",
		"confirmed", cutoff,
	).Find(&disruptions).Error

	if err != nil {
		log.Printf("[DisruptionProcessor] DB error: %v", err)
		return
	}

	for _, d := range disruptions {
		result, err := p.CoreSvc.AutoProcessDisruption(d.ID, time.Now().UTC())
		if err != nil {
			log.Printf("[DisruptionProcessor] Failed to process disruption %d: %v", d.ID, err)
			continue
		}

		// Mark disruption as processed
		now := time.Now().UTC()
		_ = p.DB.Model(&models.Disruption{}).Where("id = ?", d.ID).Update("processed_at", now).Error

		log.Printf(
			"[DisruptionProcessor] ✅ dis_%d processed: %d workers notified, %d claims, %d payouts succeeded",
			d.ID, result.WorkersNotified, result.ClaimsGenerated, result.PayoutsSucceeded,
		)
	}
}
