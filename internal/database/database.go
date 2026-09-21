package database

import (
	"fmt"
	"log"
	"SimilePro-go/internal/config"
	"SimilePro-go/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Sao_Paulo",
        cfg.DBHost,
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBName,
        cfg.DBPort,
        cfg.DBSSLMode,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    log.Println("Database connected successfully")

	// Configuração de otimização do Connection Pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("Erro ao obter a interface nativa do DB:", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(80)
	sqlDB.SetMaxIdleConns(80)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)

    // Auto-migrate models
    if err := db.AutoMigrate(
        &models.User{}, &models.Clinica{}, &models.Dentista{}, &models.Paciente{}, 
        &models.Funcionario{}, &models.Procedimento{}, &models.Agendamento{}, 
        &models.PlanoTratamento{}, &models.ItemPlano{}, &models.Evolucao{}, 
        &models.Anamnese{}, &models.Convenio{}, &models.Transacao{}, &models.Fatura{}, 
        &models.AuditLog{}, &models.CentroCusto{}, &models.PlanoConta{}, 
        &models.ConfiguracaoSistema{}, &models.Alerta{}, &models.AcessoProntuario{},
        &models.Produto{}, &models.NotaFiscalEntrada{}, &models.ItemNF{}, &models.MovimentacaoEstoque{},
        &models.CompetenciaFolha{}, &models.Holerite{}, &models.EventoHolerite{}, 
        &models.NotaFiscalServico{}, &models.ImpostoGuia{}, &models.LancamentoContabil{},
        &models.Ticket{},
    ); err != nil {
        log.Println("Warning: Auto-migration encountered an error (continuing):", err)
    }

    return db, nil
}
