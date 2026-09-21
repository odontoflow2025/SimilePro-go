package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"SimilePro-go/internal/models"
	"SimilePro-go/tests/testutils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTransacaoIsolation(t *testing.T) {
	db, cleanup, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	h := NewTransacaoHandler(db)

	clinicaA := models.Clinica{NomeFantasia: "Clinica A", CNPJ: "11.111.111/0001-11"}
	clinicaB := models.Clinica{NomeFantasia: "Clinica B", CNPJ: "22.222.222/0002-22"}
	db.Create(&clinicaA)
	db.Create(&clinicaB)

	// Lança Receita na Clínica B
	transacaoB := models.Transacao{
		ClinicaID:      clinicaB.ID,
		Descricao:      "Tratamento B",
		Valor:          5000.00,
		Tipo:           models.TipoTransacaoReceita,
		Status:         models.StatusTransacaoPago,
		DataVencimento: time.Now(),
	}
	db.Create(&transacaoB)

	t.Run("Deve isolar transacoes no GET /financeiro/transacoes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Autentica como Clinica A
		testutils.SetTestContext(c, 1, clinicaA.ID, "ADMIN")
		c.Request = httptest.NewRequest("GET", "/financeiro/transacoes", nil)

		h.FindAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var list []models.Transacao
		json.Unmarshal(w.Body.Bytes(), &list)
		
		// A clinica A não possui transações
		assert.Len(t, list, 0)
	})

	t.Run("Nao deve permitir atualizar transacao de outra clinica via PATCH", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Autentica como Clinica A tentado modificar B
		testutils.SetTestContext(c, 1, clinicaA.ID, "ADMIN")
		c.Params = []gin.Param{{Key: "id", Value: "1"}} // Transacao ID 1 (da clinica B)

		c.Request = httptest.NewRequest("PATCH", "/financeiro/transacoes/1/status", bytes.NewBufferString(`{"status":"CANCELADO"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateStatus(c)

		// Acesso negado
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
