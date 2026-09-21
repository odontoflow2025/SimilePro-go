package testutils

import (
	"context"
	"os"
	"time"

	"SimilePro-go/internal/models"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	// Set dummy env vars for all tests
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
	os.Setenv("JWT_SECRET", "test-secret")
}

// SetupTestDB initializes a PostgreSQL container and runs all migrations
func SetupTestDB() (*gorm.DB, func(), error) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, nil, err
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, nil, err
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
		&models.Funcionario{},
		&models.Transacao{},
		&models.Fatura{},
		&models.PlanoTratamento{},
		&models.ItemPlano{},
		&models.Evolucao{},
		&models.Anamnese{},
		&models.Convenio{},
		&models.CentroCusto{},
		&models.PlanoConta{},
		&models.ConfiguracaoSistema{},
		&models.Produto{},
		&models.NotaFiscalEntrada{},
		&models.ItemNF{},
		&models.MovimentacaoEstoque{},
		&models.CompetenciaFolha{},
		&models.Holerite{},
		&models.EventoHolerite{},
		&models.NotaFiscalServico{},
		&models.ImpostoGuia{},
		&models.LancamentoContabil{},
	)
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		postgresContainer.Terminate(ctx)
	}

	return db, cleanup, nil
}
