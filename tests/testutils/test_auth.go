package testutils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateTestToken creates a valid JWT for testing
func GenerateTestToken(userID uint, clinicaID uint, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       userID,
		"clinicaId": clinicaID,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

// SetTestContext injects userID and clinicaID into a gin context
func SetTestContext(c *gin.Context, userID uint, clinicaID uint) {
	c.Set("userID", userID)
	c.Set("clinicaID", clinicaID)
}
