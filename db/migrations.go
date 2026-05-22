package db

import (
	"log"

	"github.com/jamea/models"
)

func Migrate() {
	err := DB.AutoMigrate(&models.User{}, &models.Masool{}, &models.MasoolReport{}, &models.Activity{}, &models.DefaultActivity{}, &models.UserActivityLog{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	seedDefaultActivities()

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

// seedDefaultActivities inserts the built-in activities with their colors,
// skipping any that already exist so it is safe to run on every startup.
func seedDefaultActivities() {
	defaults := []models.DefaultActivity{
		{Name: "Workshop", Color: "#ef4444"},
		{Name: "Session", Color: "#f59e0b"},
		{Name: "Prayer", Color: "#3b82f6"},
		{Name: "Quiz", Color: "#10b981"},
		{Name: "Ceremony", Color: "#8b5cf6"},
		{Name: "Occasional Program", Color: "#ec4899"},
		{Name: "Meeting", Color: "#7b0827"},
		{Name: "Visits", Color: "#0ea5e9"},
		{Name: "Class", Color: "#eab308"},
		{Name: "Workshop Training", Color: "#f97316"},
		{Name: "Rally", Color: "#6366f1"},
		{Name: "Program", Color: "#06b6d4"},
		{Name: "Camp", Color: "#a855f7"},
	}

	for _, activity := range defaults {
		var count int64
		DB.Model(&models.DefaultActivity{}).Where("name = ?", activity.Name).Count(&count)
		if count == 0 {
			if err := DB.Create(&activity).Error; err != nil {
				log.Printf("Failed to seed default activity %s: %v\n", activity.Name, err)
			}
		}
	}
	log.Println("Seeded default activities")
}
