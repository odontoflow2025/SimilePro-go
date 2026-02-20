package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"odonto-flow-go/internal/handlers"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/tests/testutils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPacienteIsolation(t *testing.T) {
	// 1. Setup
	db, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	h := handlers.NewPacienteHandler(db)
	gin.SetMode(gin.TestMode)

	// Create two clinics with unique CNPJs
	clinica1 := models.Clinica{NomeFantasia: "Clinica 1", CNPJ: "11.111.111/0001-11"}
	clinica2 := models.Clinica{NomeFantasia: "Clinica 2", CNPJ: "22.222.222/0002-22"}
	db.Create(&clinica1)
	db.Create(&clinica2)

	// Create a patient in Clinic 2
	paciente := models.Paciente{
		Nome:      "Paciente Secreto",
		CPF:       "123.456.789-00",
		RG:        "RG-SECRET",
		ClinicaID: clinica2.ID,
	}
	db.Create(&paciente)

	t.Run("Should return restricted view (200 OK + flag) when Clinica 1 tries to access patient from Clinica 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Set context for Clinica 1
		testutils.SetTestContext(c, 1, clinica1.ID)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", paciente.ID)}}

		h.FindOne(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Equal(t, "ACESSO_RESTRITO", response["_privacidade"])
		assert.Contains(t, response["_mensagem"], "Histórico completo bloqueado")
		// Verify sensitive data is NOT here
		assert.Nil(t, response["rg"])
		assert.Nil(t, response["cpf"])
	})

	t.Run("Should return full data when accessing patient from same clinic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		testutils.SetTestContext(c, 2, clinica2.ID)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", paciente.ID)}}

		h.FindOne(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var p models.Paciente
		json.Unmarshal(w.Body.Bytes(), &p)
		assert.Equal(t, paciente.Nome, p.Nome)
		assert.Equal(t, paciente.RG, p.RG)
	})

	t.Run("Should return restricted data in Global Search (FindAll) for external clinic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		testutils.SetTestContext(c, 1, clinica1.ID)
		c.Request = httptest.NewRequest("GET", "/api/pacientes?search="+paciente.CPF, nil)

		h.FindAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var list []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &list)
		
		assert.Len(t, list, 1)
		assert.Equal(t, "ACESSO_RESTRITO", list[0]["_privacidade"])
		assert.Nil(t, list[0]["cpf"])
	})

	t.Run("Should unlock full data when valid Audit Protocol is provided", func(t *testing.T) {
		// 1. Generate Protocol
		wGen := httptest.NewRecorder()
		cGen, _ := gin.CreateTestContext(wGen)
		testutils.SetTestContext(cGen, 1, clinica1.ID)
		
		input := `{"pacienteId": ` + fmt.Sprintf("%d", paciente.ID) + `, "justificativa": "Necessidade clinica"}`
		cGen.Request = httptest.NewRequest("POST", "/api/pacientes/protocolo-acesso", nil)
		cGen.Request.Body = io.NopCloser(bytes.NewBufferString(input))
		cGen.Request.Header.Set("Content-Type", "application/json")

		h.GerarProtocolo(cGen)
		assert.Equal(t, http.StatusCreated, wGen.Code)
		
		var protocolResp struct {
			NumeroProtocolo string `json:"numeroProtocolo"`
		}
		json.Unmarshal(wGen.Body.Bytes(), &protocolResp)

		// 2. Use Protocol to access
		wAcc := httptest.NewRecorder()
		cAcc, _ := gin.CreateTestContext(wAcc)
		testutils.SetTestContext(cAcc, 1, clinica1.ID)
		cAcc.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", paciente.ID)}}
		cAcc.Request = httptest.NewRequest("GET", "/api/pacientes/"+fmt.Sprintf("%d", paciente.ID)+"?protocolo="+protocolResp.NumeroProtocolo, nil)

		h.FindOne(cAcc)

		assert.Equal(t, http.StatusOK, wAcc.Code)
		var p models.Paciente
		json.Unmarshal(wAcc.Body.Bytes(), &p)
		assert.Equal(t, paciente.RG, p.RG) // Full access unlocked
		assert.Empty(t, p.RG == "", "RG should be present")
	})
}
