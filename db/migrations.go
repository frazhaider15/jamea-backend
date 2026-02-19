package db

import (
	"log"

	"github.com/jamea/models"
)

func Migrate() {
	err := DB.AutoMigrate(&models.User{}, &models.Masool{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("Database migration completed")
}
