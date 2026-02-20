package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"odonto-flow-go/internal/handlers"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/tests/testutils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAgendamentoIsolation(t *testing.T) {
	// 1. Setup
	db, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	h := handlers.NewAgendamentoHandler(db)
	gin.SetMode(gin.TestMode)

	// Create two clinics
	clinica1 := models.Clinica{NomeFantasia: "Clinica A", CNPJ: "11.111.111/0001-11"}
	clinica2 := models.Clinica{NomeFantasia: "Clinica B", CNPJ: "22.222.222/0002-22"}
	db.Create(&clinica1)
	db.Create(&clinica2)

	// Create a patient in Clinic B
	paciente := models.Paciente{Nome: "Paciente B", ClinicaID: clinica2.ID}
	db.Create(&paciente)

	// Create an appointment for Patient B in Clinic B
	agendamento := models.Agendamento{
		PacienteID:     paciente.ID,
		ClinicaID:      clinica2.ID,
		DataHoraInicio: time.Now(),
		DataHoraFim:    time.Now().Add(time.Hour),
		Status:         models.StatusAgendamentoAgendado,
	}
	db.Create(&agendamento)

	t.Run("Should return 404/Forbidden when Clinic A tries to access Appointment from Clinic B", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		testutils.SetTestContext(c, 1, clinica1.ID)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", agendamento.ID)}}

		h.FindOne(c)

		// Handler should handle isolation (either 404 or 403 depending on implementation)
		assert.Contains(t, []int{http.StatusNotFound, http.StatusForbidden}, w.Code)
	})

	t.Run("Should return OK when Clinic B accesses its own Appointment", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		testutils.SetTestContext(c, 2, clinica2.ID)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", agendamento.ID)}}

		h.FindOne(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var a models.Agendamento
		json.Unmarshal(w.Body.Bytes(), &a)
		assert.Equal(t, agendamento.ID, a.ID)
	})
}
