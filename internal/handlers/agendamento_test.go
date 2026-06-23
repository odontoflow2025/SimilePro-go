package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/tests/testutils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAgendamentoBOLAIntegration(t *testing.T) {
	// 1. Instancia o Testcontainers (PostgreSQL) com auto-migrate
	db, cleanup, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup testcontainers Postgres: %v", err)
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	h := NewAgendamentoHandler(db)

	// 2. Popula o Banco de Dados com Tenants (Clínicas) Diferentes
	clinicaA := models.Clinica{NomeFantasia: "Clinica A", CNPJ: "111.111.111/0001-11"}
	clinicaB := models.Clinica{NomeFantasia: "Clinica B", CNPJ: "222.222.222/0002-22"}
	db.Create(&clinicaA)
	db.Create(&clinicaB)

	pacienteA := models.Paciente{Nome: "Paciente A", ClinicaID: clinicaA.ID}
	pacienteB := models.Paciente{Nome: "Paciente B", ClinicaID: clinicaB.ID}
	db.Create(&pacienteA)
	db.Create(&pacienteB)

	userDentistaA := models.User{Nome: "Dentista A", Email: "da@test.com", CPF: "111", TipoUsuario: "DENTISTA"}
	userDentistaB := models.User{Nome: "Dentista B", Email: "db@test.com", CPF: "222", TipoUsuario: "DENTISTA"}
	db.Create(&userDentistaA)
	db.Create(&userDentistaB)

	dentistaA := models.Dentista{UsuarioID: userDentistaA.ID, ClinicaID: clinicaA.ID, Especialidade: "Geral", CRO: "123"}
	dentistaB := models.Dentista{UsuarioID: userDentistaB.ID, ClinicaID: clinicaB.ID, Especialidade: "Ortodontia", CRO: "456"}
	db.Create(&dentistaA)
	db.Create(&dentistaB)

	agendamentoB := models.Agendamento{
		PacienteID:     pacienteB.ID,
		DentistaID:     dentistaB.ID,
		ClinicaID:      clinicaB.ID,
		DataHoraInicio: time.Now().Add(24 * time.Hour),
		DataHoraFim:    time.Now().Add(25 * time.Hour),
		Motivo:         "Revisão",
	}
	db.Create(&agendamentoB)

	// 3. Casos de Teste (OWASP API1:2023 BOLA)
	t.Run("Deve impedir que Funcionario da Clinica A acesse o agendamento da Clinica B", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Contexto injetado simulando o AuthMiddleware (Tenant Isolation)
		testutils.SetTestContext(c, 1, clinicaA.ID, "RECEPCIONISTA")
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", agendamentoB.ID)}}

		h.FindOne(c)

		// A API DEVE retornar 404 para não confirmar que o recurso existe
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Deve permitir acesso ao Agendamento quando o Tenant correto solicita", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Contexto injetado simulando o AuthMiddleware
		testutils.SetTestContext(c, 2, clinicaB.ID, "DENTISTA")
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", agendamentoB.ID)}}

		h.FindOne(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.Agendamento
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, agendamentoB.ID, response.ID)
		assert.Equal(t, agendamentoB.Motivo, response.Motivo)
	})
}
