package platform

import (
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/gorm"
)

var platformDB *gorm.DB
var platformCoreOps *services.CoreOpsService
var platformProducer *kafka.Producer

// SetDB registers DB handle for platform handlers.
func SetDB(db *gorm.DB) {
	SetDBWithProducer(db, nil)
}

// SetDBWithProducer registers DB handle and Kafka producer for platform handlers.
func SetDBWithProducer(db *gorm.DB, producer *kafka.Producer) {
	platformDB = db
	platformProducer = producer
	platformCoreOps = services.NewCoreOpsService(db, producer)
}

func hasDB() bool {
	return platformDB != nil
}
