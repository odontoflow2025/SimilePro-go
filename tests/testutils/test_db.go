package testutils

import (
	"odonto-flow-go/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes an in-memory SQLite database and runs all migrations
func SetupTestDB() (*gorm.DB, error) {
	// Use a unique name for each database instance to avoid interference between parallel tests
	// but within the same test context, "shared" allows multiple connections if needed.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Clinica{},
		&models.User{},
		&models.Dentista{},
		&models.Paciente{},
		&models.AcessoProntuario{},
		&models.Alerta{},
		&models.Agendamento{},
		&models.Procedimento{},
        &models.AuditLog{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
