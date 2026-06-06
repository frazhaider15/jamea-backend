package db

import (
	"log"
	"strings"

	"github.com/jamea/models"
)

func Migrate() {
	err := DB.AutoMigrate(&models.User{}, &models.Masool{}, &models.MasoolReport{}, &models.MasoolLocation{}, &models.Activity{}, &models.DefaultActivity{}, &models.UserActivityLog{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	seedDefaultActivities()
	normalizeMasoolsData()

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

	backfillMasoolLocations()

	log.Println("Database migration completed")
}

// backfillMasoolLocations seeds the location history for any masool that has no
// period yet, using the masool's CURRENT geography as an open-ended period that
// starts "from the beginning of time" (EffectiveFrom == ""). This covers all of
// a masool's pre-existing reports. It is correct only as long as no masool has
// already been reassigned to a different area before this runs; after the first
// real reassignment, RecordMasoolLocation splits the period instead. The pass is
// idempotent — masools that already have any location row are skipped.
func backfillMasoolLocations() {
	var masools []models.Masool
	if err := DB.Find(&masools).Error; err != nil {
		log.Printf("backfillMasoolLocations: failed to load masools: %v\n", err)
		return
	}
	created := 0
	for _, m := range masools {
		var count int64
		DB.Model(&models.MasoolLocation{}).Where("masool_id = ?", m.ID).Count(&count)
		if count > 0 {
			continue
		}
		loc := models.LocationFrom(m)
		loc.EffectiveFrom = ""
		loc.EffectiveTo = ""
		if err := DB.Create(&loc).Error; err != nil {
			log.Printf("backfillMasoolLocations: failed for masool %d: %v\n", m.ID, err)
			continue
		}
		created++
	}
	if created > 0 {
		log.Printf("Backfilled location history for %d masool(s)\n", created)
	}
}

// normalizeMasoolsData rewrites any masools.data rows that were stored in the
// legacy wrapped shape (a {"key":"data","val":[...]} entry, optionally next to
// name/module entries) into the flat shape used by CSV uploads. Idempotent —
// rows already in the flat shape are left untouched.
func normalizeMasoolsData() {
	var masools []models.Masool
	if err := DB.Find(&masools).Error; err != nil {
		log.Printf("normalizeMasoolsData: failed to load masools: %v\n", err)
		return
	}
	fixed := 0
	for i := range masools {
		normalized, changed := unwrapNestedMasoolData(masools[i].Name, masools[i].Data)
		if !changed {
			continue
		}
		masools[i].Data = normalized
		if err := DB.Save(&masools[i]).Error; err != nil {
			log.Printf("normalizeMasoolsData: failed to save masool %d: %v\n", masools[i].ID, err)
			continue
		}
		fixed++
	}
	if fixed > 0 {
		log.Printf("Normalized data shape for %d masool(s)\n", fixed)
	}
}

func unwrapNestedMasoolData(name string, data []models.MasoolData) ([]models.MasoolData, bool) {
	var inner []models.MasoolData
	found := false

	for _, entry := range data {
		if !strings.EqualFold(strings.TrimSpace(entry.Key), "data") {
			continue
		}
		switch v := entry.Val.(type) {
		case []interface{}:
			for _, item := range v {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				k, _ := m["key"].(string)
				if strings.TrimSpace(k) == "" {
					continue
				}
				inner = append(inner, models.MasoolData{Key: k, Val: m["val"]})
			}
			found = true
		case []models.MasoolData:
			inner = v
			found = true
		}
		break
	}

	if !found {
		return data, false
	}

	normalized := []models.MasoolData{{Key: "Name", Val: name}}
	for _, entry := range inner {
		kl := strings.ToLower(strings.TrimSpace(entry.Key))
		if kl == "" || kl == "name" || kl == "module" {
			continue
		}
		normalized = append(normalized, entry)
	}
	return normalized, true
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
