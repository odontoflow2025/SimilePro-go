package main

import (
	"fmt"
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"
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

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}
    
    fmt.Printf("Users in DB:\n")
    for _, u := range users {
        fmt.Printf("- %s (Email: %s) -> ClinicaID: %d\n", u.Nome, u.Email, u.ClinicaID)
    }

	var clinicas []models.Clinica
	if err := db.Find(&clinicas).Error; err != nil {
		log.Fatalf("Failed to fetch clinics: %v", err)
	}
    fmt.Printf("\nClinics in DB:\n")
    for _, c := range clinicas {
        fmt.Printf("- %s (ID: %d)\n", c.NomeFantasia, c.ID)
    }
}
