package core

import "gorm.io/gorm"

var coreDB *gorm.DB

// SetDB registers DB handle for core handlers.
func SetDB(db *gorm.DB) {
	coreDB = db
}

func hasDB() bool {
	return coreDB != nil
}
