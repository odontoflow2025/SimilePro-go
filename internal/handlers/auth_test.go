package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/tests/testutils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := testutils.SetupTestDB()
	h := NewAuthHandler(db, "test-secret")

	t.Run("Should block ADMIN_TOTAL and default to RECEPCIONISTA", func(t *testing.T) {
		input := RegisterInput{
			Nome: "Fake Admin",
			Email: "fake@admin.com",
			Senha: "password123",
			CPF: "000.000.001-01",
			TipoUsuario: "ADMIN_TOTAL",
		}
		body, _ := json.Marshal(input)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
		
		h.Register(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var user models.User
		db.Where("email = ?", input.Email).First(&user)
		assert.Equal(t, "RECEPCIONISTA", user.TipoUsuario)
	})

	t.Run("Should allow DENTISTA", func(t *testing.T) {
		input := RegisterInput{
			Nome: "Dentista 1",
			Email: "dentista@teste.com",
			Senha: "password123",
			CPF: "000.000.001-02",
			TipoUsuario: "DENTISTA",
		}
		body, _ := json.Marshal(input)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
		
		h.Register(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var user models.User
		db.Where("email = ?", input.Email).First(&user)
		assert.Equal(t, "DENTISTA", user.TipoUsuario)
	})
}
