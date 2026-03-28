package testutils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateTestToken creates a valid JWT for testing
func GenerateTestToken(userID uint, clinicaID uint, role string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         userID,
		"clinicaId":   clinicaID,
		"tipoUsuario": role,
		"exp":         time.Now().Add(time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

// SetTestContext injects userID, clinicaID and userRole into a gin context
func SetTestContext(c *gin.Context, userID uint, clinicaID uint, role string) {
	c.Set("userID", userID)
	c.Set("clinicaID", clinicaID)
	c.Set("userRole", role)
}
