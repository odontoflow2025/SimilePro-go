package middleware

import (
	"net/http"
	"net/http/httptest"
	"SimilePro-go/internal/models"
	"SimilePro-go/tests/testutils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup, _ := testutils.SetupTestDB()
	defer cleanup()

	// Create test users with different roles
	admin := models.User{Nome: "Admin", Email: "admin@rbac.com", CPF: "111.111.111-11", TipoUsuario: "ADMIN_TOTAL"}
	dentista := models.User{Nome: "Dentista", Email: "dentista@rbac.com", CPF: "222.222.222-22", TipoUsuario: "DENTISTA"}
	db.Create(&admin)
	db.Create(&dentista)

	// Add Dentista residency mapping for the clinic
	db.Create(&models.Dentista{
		UsuarioID: dentista.ID,
		ClinicaID: 1,
	})

	t.Run("Should allow ADMIN_TOTAL to access admin-only route", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		testutils.SetTestContext(c, admin.ID, 1, "ADMIN_TOTAL")

		handler := RequireRole(db, "ADMIN_TOTAL")
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("Should deny DENTISTA from accessing admin-only route", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		testutils.SetTestContext(c, dentista.ID, 1, "DENTISTA")

		handler := RequireRole(db, "ADMIN_TOTAL")
		handler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("Should allow multiple roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		testutils.SetTestContext(c, dentista.ID, 1, "DENTISTA")

		handler := RequireRole(db, "ADMIN_TOTAL", "DENTISTA")
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
	})
}
