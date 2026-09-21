package handlers

import (
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

func TestContabilidadeDREIsolation(t *testing.T) {
	db, cleanup, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	h := NewContabilidadeHandler(db)

	clinicaA := models.Clinica{NomeFantasia: "Clinica A Contabil", CNPJ: "11.111.111/0001-11"}
	clinicaB := models.Clinica{NomeFantasia: "Clinica B Contabil", CNPJ: "22.222.222/0002-22"}
	db.Create(&clinicaA)
	db.Create(&clinicaB)

	// Lança Receita PAGA na Clínica B para o DRE aparecer
	data := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	transacaoB := models.Transacao{
		ClinicaID:      clinicaB.ID,
		Descricao:      "Tratamento B",
		Valor:          50000.00,
		Tipo:           models.TipoTransacaoReceita,
		Status:         models.StatusTransacaoPago,
		DataVencimento: data,
		DataPagamento:  &data,
	}
	db.Create(&transacaoB)

	t.Run("O DRE da Clinica A deve ser 0 pois nao tem transacoes (BOLA prevention)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Autentica como Clinica A
		testutils.SetTestContext(c, 1, clinicaA.ID, "ADMIN")
		c.Request = httptest.NewRequest("GET", "/contabilidade/dre/mensal?mes=6&ano=2026", nil)

		h.GetDREMensal(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var dre map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &dre)
		
		// Esperamos 0 em grossRevenue para a Clinica A
		assert.Equal(t, 0.0, dre["grossRevenue"])
	})

	t.Run("O DRE da Clinica B deve refletir a receita dela isoladamente", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Autentica como Clinica B
		testutils.SetTestContext(c, 2, clinicaB.ID, "ADMIN")
		c.Request = httptest.NewRequest("GET", "/contabilidade/dre/mensal?mes=6&ano=2026", nil)

		h.GetDREMensal(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var dre map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &dre)
		
		// Esperamos 50000 em grossRevenue para a Clinica B
		assert.Equal(t, 50000.0, dre["grossRevenue"])
	})
}
