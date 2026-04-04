package core

import (
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/gorm"
)

var coreDB *gorm.DB
var coreOps *services.CoreOpsService
var coreProducer *kafka.Producer

// SetDB registers DB handle for core handlers.
func SetDB(db *gorm.DB) {
	SetDBWithProducer(db, nil)
}

// SetDBWithProducer registers DB handle and Kafka producer for core handlers.
func SetDBWithProducer(db *gorm.DB, producer *kafka.Producer) {
	coreDB = db
	coreProducer = producer
	coreOps = services.NewCoreOpsService(db, producer)
}

func hasDB() bool {
	return coreDB != nil && coreOps != nil
}
