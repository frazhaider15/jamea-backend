package models

type Activity struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Module     Module   `gorm:"not null;index:idx_module_month_year,unique" json:"module"`
	Month      int      `gorm:"not null;index:idx_module_month_year,unique" json:"month"`
	Year       int      `gorm:"not null;index:idx_module_month_year,unique" json:"year"`
	Activities []string `gorm:"type:jsonb;serializer:json" json:"activities"`
}
