package middleware

import (
	"net/http"
	"net/http/httptest"
	"SimilePro-go/internal/config"
	"SimilePro-go/tests/testutils"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Change to project root so os.ReadFile("public.pem") inside middleware works
	cwd, _ := os.Getwd()
	os.Chdir("../../")
	defer os.Chdir(cwd)

	gin.SetMode(gin.TestMode)
	
	// Load config for secret
	cfg, _ := config.LoadConfig()
	secret := cfg.JWTSecret

	t.Run("Should pass with valid token and set context", func(t *testing.T) {
		token, _ := testutils.GenerateTestToken(1, 10, "DENTISTA", secret)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		handler := AuthMiddleware(secret)
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		userID, _ := c.Get("userID")
		clinicaID, _ := c.Get("clinicaID")
		userRole, _ := c.Get("userRole")
		
		assert.Equal(t, uint(1), userID)
		assert.Equal(t, uint(10), clinicaID)
		assert.Equal(t, "DENTISTA", userRole)
	})

	t.Run("Should fail with missing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		handler := AuthMiddleware(secret)
		handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("Should fail with invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		handler := AuthMiddleware(secret)
		handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})
}
