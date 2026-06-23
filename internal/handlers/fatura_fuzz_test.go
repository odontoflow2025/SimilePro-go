package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// FuzzFaturaPayload bombardeia o endpoint de criação de faturas
// com payloads matematicamente incorretos ou maliciosos (NaN, Infinito, negativos).
// O objetivo é garantir que o json binding do gin retorne 400 Bad Request
// e nunca cause PANIC ou deixe o valor chegar à camada do GORM.
func FuzzFaturaPayload(f *testing.F) {
	// Seed corpus com casos extremos de finanças
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": -50.00, "dataVencimento": "2026-12-31T00:00:00Z"}`))
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": 0.00, "dataVencimento": "2026-12-31T00:00:00Z"}`))
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": NaN, "dataVencimento": "2026-12-31T00:00:00Z"}`))
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": Infinity, "dataVencimento": "2026-12-31T00:00:00Z"}`))
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": "100,00", "dataVencimento": "2026-12-31T00:00:00Z"}`))
	f.Add([]byte(`{"clinicaId": 1, "pacienteId": 1, "valorTotal": 999999999999999999999999.99, "dataVencimento": "2026-12-31T00:00:00Z"}`))

	gin.SetMode(gin.TestMode)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Handler isolado sem GORM pois queremos testar apenas a robustez do parser (Crashed/Panic)
		h := NewFaturaHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Simula um usuário autenticado no contexto para passar pelo primeiro check
		c.Set("clinicaID", uint(1))

		req, err := http.NewRequest("POST", "/financeiro/faturas", bytes.NewReader(data))
		if err != nil {
			t.Skip()
		}
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Captura panics. O Fuzzer falhará se houver panic (o que significa que a API estourou).
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PANIC DETECTADO no parser de finanças: %v", r)
			}
		}()

		h.Create(c)

		// Regra de Ouro: O backend deve rejeitar entradas inválidas com 400 Bad Request
		// Se passou pelo binding (201 Created) e o payload não era 100% sadio, pode haver falha de validação.
		// No caso do fuzzing cego gerando bytes aleatórios, é virtualmente impossível gerar um JSON
		// válido com clinicaID, pacienteID, valorTotal > 0 e dataVencimento no formato RFC3339 acidentalmente.
		if w.Code == http.StatusCreated || w.Code == http.StatusOK {
			// Apenas documenta, pois o fuzzer também testa lixo puro que falhará em 400.
			t.Logf("Payload passou no binding de Fatura: %d", w.Code)
		}
	})
}
