package core

import (
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/gorm"
)

var coreDB *gorm.DB
var coreOps *services.CoreOpsService

// SetDB registers DB handle for core handlers.
func SetDB(db *gorm.DB) {
	coreDB = db
	coreOps = services.NewCoreOpsService(db)
}

func hasDB() bool {
	return coreDB != nil && coreOps != nil
}
