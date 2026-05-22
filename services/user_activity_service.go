package services

import (
	"log"

	"github.com/jamea/db"
	"github.com/jamea/models"
)

// LogUserActivity records an action performed by a user (scoped by module).
// It is best-effort: a failure is logged but never blocks the calling request.
func LogUserActivity(module models.Module, action, description string) {
	entry := models.UserActivityLog{
		Module:      module,
		Action:      action,
		Description: description,
	}
	if err := db.DB.Create(&entry).Error; err != nil {
		log.Printf("failed to record user activity (%s): %v\n", action, err)
	}
}

// GetLastUserActivities returns the most recent activity log entries,
// newest first. If module is empty, entries across all modules are returned.
func GetLastUserActivities(module models.Module, limit int) ([]models.UserActivityLog, error) {
	var logs []models.UserActivityLog
	query := db.DB.Order("created_at DESC")
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}
