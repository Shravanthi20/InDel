package models

import "time"

type ActivePolicy struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"uniqueIndex"`
	PolicyID  uint
	Zone      string
	StartedAt time.Time
	UpdatedAt time.Time
}

func (ActivePolicy) TableName() string {
	return "active_policies"
}
