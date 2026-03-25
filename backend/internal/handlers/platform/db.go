package platform

import "gorm.io/gorm"

var platformDB *gorm.DB

// SetDB registers DB handle for platform handlers.
func SetDB(db *gorm.DB) {
	platformDB = db
}

func hasDB() bool {
	return platformDB != nil
}
