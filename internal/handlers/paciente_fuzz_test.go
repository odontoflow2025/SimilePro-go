package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"odonto-flow-go/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// FuzzCreatePaciente aplica a técnica de Fuzzing no Handler de Criação de Pacientes.
// O objetivo é testar se o Gin/GORM ou nossa lógica entra em colapso (Panic) quando
// recebe bytes e payloads mutados aleatoriamente, garantindo resiliência total a inputs.
func FuzzCreatePaciente(f *testing.F) {
	// 1. Fornecer uma base de dados conhecida ("seed corpus") para o fuzzer mutar.
	f.Add([]byte(`{"nome":"João","telefonePrincipal":"11999999999","clinicaId":1}`)) // Válido
	f.Add([]byte(`{"nome":"","telefonePrincipal":""}`))                              // Inválido (faltando required)
	f.Add([]byte(`{malformed json]`))                                                // Sintaxe quebrada
	f.Add([]byte(`{"nome": 123}`))                                                   // Tipagem errada
	f.Add([]byte(`{"nome":"Hacker","role":"admin","tenant_id":2}`))                  // Mass Assignment (Payload excessivo)
	f.Add([]byte(`null`))                                                            // Body nulo
	f.Add([]byte(`[{"nome":"João"}]`))                                               // Array no lugar de objeto

	// Configuração do ambiente global apenas uma vez
	gin.SetMode(gin.TestMode)

	// 2. A função Fuzz executará milhares de vezes com dados mutados ("data")
	f.Fuzz(func(t *testing.T, data []byte) {
		// Inicializamos um banco in-memory leve para garantir que o fluxo vá até o fim
		// sem mocks complexos que poderiam esconder falhas de integração.
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Skip("Falha ao inicializar o banco in-memory")
		}
		
		// Opcional: auto migrate apenas se precisar validar persistência.
		// Mesmo sem as tabelas, o GORM lidará retornando erro (o que é válido e testável aqui).
		_ = db.AutoMigrate(&models.Paciente{})

		h := NewPacienteHandler(db)

		// Mock do contexto do Gin e do ResponseWriter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Simulação de requisição recebendo o byte[] caótico do Fuzzer
		req, err := http.NewRequest(http.MethodPost, "/pacientes", bytes.NewReader(data))
		if err != nil {
			t.Skip("Requisição inválida gerada pelo fuzzer, ignorando iteração.")
		}
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Injeta o que o Middleware de Autenticação injetaria num fluxo real
		c.Set("clinicaID", uint(1))

		// 3. O Foco da Asserção: 
		// Simplesmente chamar h.Create(c) dentro do `f.Fuzz` é suficiente para a engine de teste do Go.
		// Se essa chamada causar qualquer tipo de *panic* (ex: tentar acessar um pointer nil, slice out of bounds),
		// o teste falhará imediatamente, nos avisando da vulnerabilidade.
		h.Create(c)

		// 4. Verificações extras de integridade (opcional para Fuzzing, mas recomendado)
		// Como enviamos dados bizarros, a API deve sempre retornar um status controlado (400, 422, 500 ou 201), nunca crashear.
		if w.Code != http.StatusBadRequest && 
		   w.Code != http.StatusInternalServerError && 
		   w.Code != http.StatusCreated {
			// Se retornar algo fora desse escopo (ex: 200 OK sem criar nada, ou 401 sem auth handler), algo pode estar inconsistente.
			t.Logf("Aviso: status code inusitado retornado: %d", w.Code)
		}
	})
}
