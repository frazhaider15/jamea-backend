package db

import (
	"log"

	"github.com/jamea/models"
)

func Migrate() {
	err := DB.AutoMigrate(&models.User{}, &models.Masool{}, &models.MasoolReport{}, &models.Activity{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	var count int64
	DB.Model(&models.Masool{}).Where("id = ?", 40).Count(&count)
	if count == 0 {
		DB.Create(&models.Masool{
			ID:     40,
			Name:   "Sample Masool",
			Module: models.ModuleAms,
		})
	}

	DB.Model(&models.MasoolReport{}).Where("masool_id = ?", 40).Count(&count)

	reports := []models.MasoolReport{
		{
			MasoolID: 40,
			Module:   models.ModuleAms,
			Month:    "2026_01",
			Data: []models.MasoolData{
				{Key: "Name", Val: "Faraz"},
				{Key: "Activities", Val: []string{"quiz", "class", "program", "meeting"}},
				{Key: "Remarks", Val: []string{"40 logon ny shirkat ki", "20 logon ny shirkat ki", "10 logon ny shirkat ki", "15 logon ny shirkat ki"}},
			},
		},
		{
			MasoolID: 40,
			Module:   models.ModuleAms,
			Month:    "2026_02",
			Data: []models.MasoolData{
				{Key: "Name", Val: "Faraz"},
				{Key: "Activities", Val: []string{"quiz", "class"}},
				{Key: "Remarks", Val: []string{"20 logon ny shirkat ki", "10 logon ny shirkat ki"}},
			},
		},
		{
			MasoolID: 40,
			Module:   models.ModuleAms,
			Month:    "2026_03",
			Data: []models.MasoolData{
				{Key: "Name", Val: "Faraz"},
				{Key: "Activities", Val: []string{"program", "meeting"}},
				{Key: "Remarks", Val: []string{"15 logon ny shirkat ki", "10 logon ny shirkat ki"}},
			},
		},
	}
	for _, report := range reports {
		DB.Create(&report)
	}
	log.Println("Seeded MasoolReports for masool_id 40")

	log.Println("Database migration completed")
}
