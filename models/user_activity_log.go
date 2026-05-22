package models

import "time"

// UserActivityLog records an action performed by a user, scoped by module.
type UserActivityLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Module      Module    `gorm:"index" json:"module"`
	Action      string    `gorm:"not null" json:"action"`
	Description string    `gorm:"not null" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
