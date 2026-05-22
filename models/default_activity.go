package models

type DefaultActivity struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"not null;uniqueIndex" json:"name"`
	Color string `gorm:"not null" json:"color"`
}
