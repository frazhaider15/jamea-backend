package db

import (
	"log"

	"github.com/jamea/models"
)

func Seed() {
	users := []models.User{
		{
			Name:     "AMS Admin",
			Email:    "ams_admin@gmail.com",
			Password: "admin123",
			Module:   models.ModuleAms,
		},
		{
			Name:     "TBUM Admin",
			Email:    "tbum_admin@gmail.com",
			Password: "admin123",
			Module:   models.ModuleTbum,
		},
		{
			Name:     "PTBUM Admin",
			Email:    "ptbum_admin@gmail.com",
			Password: "admin123",
			Module:   models.ModulePtbum,
		},
	}

	for _, u := range users {
		var existing models.User
		if err := DB.Where("email = ?", u.Email).First(&existing).Error; err != nil {
			// Assuming record not found, create it
			if err := DB.Create(&u).Error; err != nil {
				log.Printf("Failed to seed user %s: %v\n", u.Email, err)
			} else {
				log.Printf("Seeded user: %s\n", u.Email)
			}
		}
	}
	log.Println("Seeding completed")
}
