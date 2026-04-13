package models

import "time"

type Batch struct {
	BatchID          string     `gorm:"primaryKey;column:batch_id"`
	ZoneLevel        string     `gorm:"column:zone_level"`
	FromCity         string     `gorm:"column:from_city"`
	ToCity           string     `gorm:"column:to_city"`
	TotalWeight      float64    `gorm:"column:total_weight"`
	OrderCount       int        `gorm:"column:order_count"`
	Status           string     `gorm:"column:status"`
	PickupCode       string     `gorm:"column:pickup_code"`
	DeliveryCode     string     `gorm:"column:delivery_code"`
	PickupUserID     *uint      `gorm:"column:pickup_user_id"`
	PickupTime       *time.Time `gorm:"column:pickup_time"`
	DeliveryTime     *time.Time `gorm:"column:delivery_time"`
	BatchEarningINR  float64    `gorm:"column:batch_earning_inr"`
	EarningsPosted   bool       `gorm:"column:earnings_posted"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

type BatchOrder struct {
	OrderID          string     `gorm:"primaryKey;column:order_id"`
	BatchID          string     `gorm:"primaryKey;column:batch_id"`
	UserID           *uint      `gorm:"column:user_id"`
	Status           string     `gorm:"column:status"`
	PickupTime       *time.Time `gorm:"column:pickup_time"`
	DeliveryTime     *time.Time `gorm:"column:delivery_time"`
	DeliveryAddress  string     `gorm:"column:delivery_address"`
	ContactName      string     `gorm:"column:contact_name"`
	ContactPhone     string     `gorm:"column:contact_phone"`
	Weight           float64    `gorm:"column:weight"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}