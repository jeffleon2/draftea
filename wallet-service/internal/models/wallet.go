package models

import "time"

type Wallet struct {
	ID        string  `gorm:"primaryKey"`
	UserID    string  `gorm:"index;not null"`
	Balance   float64 `gorm:"not null"`
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
