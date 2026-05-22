package services

import (
	"github.com/jamea/db"
	"github.com/jamea/models"
)

func SaveActivities(activity models.Activity) (models.Activity, error) {
	var existing models.Activity
	err := db.DB.Where("module = ? AND month = ? AND year = ?", activity.Module, activity.Month, activity.Year).First(&existing).Error

	if err == nil {
		// Update existing record
		existing.Activities = activity.Activities
		if err := db.DB.Save(&existing).Error; err != nil {
			return existing, err
		}
		return existing, nil
	}

	// Create new record
	if err := db.DB.Create(&activity).Error; err != nil {
		return activity, err
	}
	return activity, nil
}

func GetActivities(module models.Module, month, year int) ([]models.Activity, error) {
	var activities []models.Activity
	query := db.DB.Where("module = ?", module)
	if month != 0 {
		query = query.Where("month = ?", month)
	}
	if year != 0 {
		query = query.Where("year = ?", year)
	}
	err := query.Find(&activities).Error
	return activities, err
}
func DeleteActivitiesByModule(module models.Module) error {
	return db.DB.Where("module = ?", module).Delete(&models.Activity{}).Error
}
