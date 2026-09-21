package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"SimilePro-go/internal/handlers"
	"SimilePro-go/internal/models"
	"SimilePro-go/tests/testutils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClinicIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanup()
	h := handlers.NewPacienteHandler(db)

	clinica1 := models.Clinica{NomeFantasia: "Clinica 1", CNPJ: "11.111.111/0001-11"}
	clinica2 := models.Clinica{NomeFantasia: "Clinica 2", CNPJ: "22.222.222/0002-22"}
	db.Create(&clinica1)
	db.Create(&clinica2)

	t.Run("Should ignore clinic ID from body and use auth context ID (Clinic Isolation Support)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Authenticated as Clinica 1
		testutils.SetTestContext(c, 1, clinica1.ID, "RECEPCIONISTA")
		
		// Attempting to create patient in Clinica 2
		body := map[string]interface{}{
			"nome": "Paciente Intrusion",
			"clinicaId": clinica2.ID, // User Tries to set Clinica 2
			"telefonePrincipal": "999999999",
		}
		bodyBytes, _ := json.Marshal(body)
		
		c.Request = httptest.NewRequest("POST", "/api/pacientes", bytes.NewBuffer(bodyBytes))

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var p models.Paciente
		json.Unmarshal(w.Body.Bytes(), &p)
		
		// Verified: it forced Clinica 1 (Auth context) regardless of body input
		assert.Equal(t, clinica1.ID, p.ClinicaID) 
		assert.NotEqual(t, clinica2.ID, p.ClinicaID)
	})

	t.Run("Should prevent unauthorized employee dismissal (Cross-Clinic IDOR)", func(t *testing.T) {
		hFunc := handlers.NewFuncionarioHandler(db)
		
		// Create employee in Clinica 2
		emp2 := models.Funcionario{
			UsuarioID: 999,
			ClinicaID: clinica2.ID,
			Cargo: "DENTISTA",
		}
		db.Create(&emp2)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Authenticated as Clinica 1
		testutils.SetTestContext(c, 1, clinica1.ID, "ADMIN_GERENCIAL")
		
		c.Params = []gin.Param{{Key: "id", Value: "1"}} // Note: SetupTestDB usually starts with ID 1
		
		// Try to dismiss employee from Clinica 2
		body := map[string]interface{}{
			"dataDemissao": "2024-03-24T12:00:00Z",
			"motivoDemissao": "Hacker attempt",
		}
		bodyBytes, _ := json.Marshal(body)
		c.Request = httptest.NewRequest("PATCH", "/api/funcionarios/1/demitir", bytes.NewBuffer(bodyBytes))

		hFunc.Demitir(c)

		// Expected Forbidden as it belongs to another clinic
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Should enforce isolation in Financial Transactions", func(t *testing.T) {
		hTrans := handlers.NewTransacaoHandler(db)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		testutils.SetTestContext(c, 1, clinica1.ID, "FINANCEIRO")
		
		body := map[string]interface{}{
			"descricao": "Fraudulent Income",
			"valor": 100000.0,
			"clinicaId": clinica2.ID, // Try to attribute to other clinic
			"tipo": "RECEITA",
			"dataVencimento": "2024-03-24T12:00:00Z",
		}
		bodyBytes, _ := json.Marshal(body)
		c.Request = httptest.NewRequest("POST", "/api/financeiro/transacoes", bytes.NewBuffer(bodyBytes))
		
		hTrans.Create(c)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		var tr models.Transacao
		json.Unmarshal(w.Body.Bytes(), &tr)
		
		assert.Equal(t, clinica1.ID, tr.ClinicaID)
	})
}
