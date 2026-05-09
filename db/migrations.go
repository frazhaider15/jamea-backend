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

	// Seed previous months reports for new Masools
	newMasools := []struct {
		ID   uint
		Name string
	}{
		{4, "Yasir"},
		{5, "Mehran"},
		{6, "Ali"},
		{7, "Faheem"},
		{8, "Faraz"},
		{9, "Khan"},
		{10, "Ammar"},
		{11, "Ahsan"},
		{12, "Jarar"},
		{13, "Hassan"},
		{14, "Zafar"},
	}

	months := []string{"2026_01", "2026_02", "2026_03", "2026_04"}

	for _, m := range newMasools {
		var exists int64
		DB.Model(&models.Masool{}).Where("id = ?", m.ID).Count(&exists)
		if exists == 0 {
			DB.Create(&models.Masool{
				ID:     m.ID,
				Name:   m.Name,
				Module: models.ModuleAms,
			})
		}

		for _, month := range months {
			var reportCount int64
			DB.Model(&models.MasoolReport{}).Where("masool_id = ? AND month = ?", m.ID, month).Count(&reportCount)
			if reportCount == 0 {
				DB.Create(&models.MasoolReport{
					MasoolID: m.ID,
					Module:   models.ModuleAms,
					Month:    month,
					Data: []models.MasoolData{
						{Key: "Name", Val: m.Name},
						{Key: "Activities", Val: []string{"quiz", "class", "program", "meeting"}},
						{Key: "Remarks", Val: []string{"shirkat achi rahi", "attendance behtar thi", "maqsad hasil hua", "mazeed behtari ki gunjayish hai"}},
					},
				})
			}
		}
	}
	log.Println("Seeded previous months reports for Masools 4-14")

	log.Println("Database migration completed")
}
