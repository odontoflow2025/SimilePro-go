package handlers

import (
	"bytes"
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

func TestFaturaIsolationAndMutability(t *testing.T) {
	db, cleanup, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	h := NewFaturaHandler(db)

	// Setup Clinics
	clinicaA := models.Clinica{NomeFantasia: "Clinica Financeira A", CNPJ: "10.000.000/0001-00"}
	clinicaB := models.Clinica{NomeFantasia: "Clinica Financeira B", CNPJ: "20.000.000/0002-00"}
	db.Create(&clinicaA)
	db.Create(&clinicaB)

	// Setup Faturas
	faturaA_Pendente := models.Fatura{
		ClinicaID:      clinicaA.ID,
		ValorTotal:     500.00,
		Status:         models.StatusFaturaPendente,
		DataVencimento: time.Now().Add(48 * time.Hour),
	}
	db.Create(&faturaA_Pendente)

	faturaB_Paga := models.Fatura{
		ClinicaID:      clinicaB.ID,
		ValorTotal:     1200.00,
		Status:         models.StatusFaturaPaga,
		DataVencimento: time.Now().Add(-24 * time.Hour),
	}
	db.Create(&faturaB_Paga)

	t.Run("Deve impedir que Clinica A altere fatura da Clinica B (BOLA)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Contexto como Clinica A
		testutils.SetTestContext(c, 1, clinicaA.ID, "RECEPCIONISTA")
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", faturaB_Paga.ID)}}

		input := `{"valorTotal": 10.00, "dataVencimento": "2026-12-31T00:00:00Z"}`
		req, _ := http.NewRequest("PUT", "/financeiro/faturas/"+fmt.Sprintf("%d", faturaB_Paga.ID), bytes.NewBufferString(input))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		h.Update(c)

		// Restrição BOLA para alteração financeira deve retornar 403 (Forbidden) conforme o código do Handler.
		// Observação: Na busca (FindOne) geralmente é 404 para ocultar existência, 
		// mas em comandos explícitos de alteração de ID conhecido via URL, 403 Forbidden é emitido pelo ownership check.
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Deve impedir alteração de valor de fatura PAGA da propria Clinica", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Contexto como Clinica B
		testutils.SetTestContext(c, 2, clinicaB.ID, "ADMIN")
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", faturaB_Paga.ID)}}

		input := `{"valorTotal": 1500.00, "dataVencimento": "2026-12-31T00:00:00Z"}`
		req, _ := http.NewRequest("PUT", "/financeiro/faturas/"+fmt.Sprintf("%d", faturaB_Paga.ID), bytes.NewBufferString(input))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		h.Update(c)

		// Deve retornar 409 Conflict devido ao status PAGO
		assert.Equal(t, http.StatusConflict, w.Code)
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "não podem ser alteradas")
	})

	t.Run("Deve permitir alteração de fatura PENDENTE da propria Clinica", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Contexto como Clinica A
		testutils.SetTestContext(c, 1, clinicaA.ID, "ADMIN")
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", faturaA_Pendente.ID)}}

		input := `{"valorTotal": 650.00, "dataVencimento": "2026-12-31T00:00:00Z"}`
		req, _ := http.NewRequest("PUT", "/financeiro/faturas/"+fmt.Sprintf("%d", faturaA_Pendente.ID), bytes.NewBufferString(input))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var updatedFatura models.Fatura
		h.DB.First(&updatedFatura, faturaA_Pendente.ID)
		assert.Equal(t, 650.00, updatedFatura.ValorTotal)
	})
}
