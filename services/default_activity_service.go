package services

import (
	"errors"

	"github.com/jamea/db"
	"github.com/jamea/models"
	"gorm.io/gorm"
)

// DefaultActivityColor is used when a new default activity is added without a color.
const DefaultActivityColor = "#6b7280"

func GetDefaultActivities() ([]models.DefaultActivity, error) {
	var activities []models.DefaultActivity
	err := db.DB.Order("id ASC").Find(&activities).Error
	return activities, err
}

func AddDefaultActivity(activity models.DefaultActivity) (models.DefaultActivity, error) {
	if activity.Color == "" {
		activity.Color = DefaultActivityColor
	}

	var existing models.DefaultActivity
	err := db.DB.Where("name = ?", activity.Name).First(&existing).Error
	if err == nil {
		return existing, errors.New("activity already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return activity, err
	}

	if err := db.DB.Create(&activity).Error; err != nil {
		return activity, err
	}
	return activity, nil
}
