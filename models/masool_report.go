package models

type MasoolReport struct {
	ID       uint         `gorm:"primaryKey" json:"id"`
	MasoolID uint         `gorm:"not null;index" json:"masool_id"`
	Module   Module       `gorm:"not null;index" json:"module"`
	Month    string       `gorm:"not null;index;default:''" json:"month"`
	Data     []MasoolData `gorm:"type:jsonb;serializer:json" json:"data"`
}
