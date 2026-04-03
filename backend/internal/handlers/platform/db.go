package platform

import (
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/gorm"
)

var platformDB *gorm.DB
var platformCoreOps *services.CoreOpsService

// SetDB registers DB handle for platform handlers.
func SetDB(db *gorm.DB) {
	platformDB = db
	platformCoreOps = services.NewCoreOpsService(db)
}

func hasDB() bool {
	return platformDB != nil
}
