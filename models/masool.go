package models

type MasoolData struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type Masool struct {
	ID     uint         `gorm:"primaryKey" json:"id"`
	Name   string       `gorm:"not null" json:"name"`
	Module Module       `gorm:"not null;index" json:"module"` // Added module mapping
	Data   []MasoolData `gorm:"type:jsonb;serializer:json" json:"data"`
}
