package database

import (
	"fmt"
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
        cfg.DBHost,
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBName,
        cfg.DBPort,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    log.Println("Database connected successfully")

    // Auto-migrate models
    if err := db.AutoMigrate(&models.User{}, &models.Clinica{}, &models.Dentista{}, &models.Paciente{}, &models.Funcionario{}, &models.Procedimento{}, &models.Agendamento{}, &models.PlanoTratamento{}, &models.ItemPlano{}, &models.Evolucao{}, &models.Anamnese{}, &models.Convenio{}, &models.Transacao{}, &models.Fatura{}, &models.AuditLog{}, &models.CentroCusto{}, &models.PlanoConta{}, &models.ConfiguracaoSistema{}, &models.Alerta{}, &models.AcessoProntuario{}); err != nil {
        log.Println("Warning: Auto-migration encountered an error (continuing):", err)
    }

    return db, nil
}
