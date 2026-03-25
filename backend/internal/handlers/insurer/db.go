package insurer

import "gorm.io/gorm"

var insurerDB *gorm.DB

// SetDB registers DB handle for insurer handlers.
func SetDB(db *gorm.DB) {
	insurerDB = db
}

func hasDB() bool {
	return insurerDB != nil
}
