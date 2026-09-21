package main

import (
	"log"
	"SimilePro-go/internal/config"
	"SimilePro-go/internal/database"
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

    // Force raw SQL to avoid Gorm struct empty value skipping or other silent failures
	result := db.Exec("UPDATE users SET clinica_id = 1 WHERE clinica_id = 0 OR clinica_id IS NULL")
    if result.Error != nil {
        log.Fatalf("SQL Error: %v", result.Error)
    }

	log.Printf("Successfully updated %d users to clinica_id = 1", result.RowsAffected)
}
