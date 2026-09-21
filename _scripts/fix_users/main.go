package main

import (
	"log"
	"SimilePro-go/internal/config"
	"SimilePro-go/internal/database"
	"SimilePro-go/internal/models"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	var clinica models.Clinica
	if err := db.First(&clinica).Error; err != nil {
		log.Fatalf("No clinic found in DB: %v", err)
	}

    // Update all users with ClinicaID = 0 to the first clinic's ID
    if err := db.Model(&models.User{}).Where("clinica_id = 0").Update("clinica_id", clinica.ID).Error; err != nil {
        log.Fatalf("Failed to update users: %v", err)
    }

    log.Printf("Successfully linked all unassigned users to Clinic ID %d (%s)", clinica.ID, clinica.NomeFantasia)
}
