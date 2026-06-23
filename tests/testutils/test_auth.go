package testutils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// getProjectRoot walks up the directory tree to find the go.mod file
func getProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// GenerateTestToken creates a valid JWT for testing using RS256
func GenerateTestToken(userID uint, clinicaID uint, role string, _ string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":         float64(userID),
		"clinicaId":   float64(clinicaID),
		"tipoUsuario": role,
		"exp":         float64(time.Now().Add(time.Hour).Unix()),
	})

	privateKeyBytes, err := os.ReadFile(filepath.Join(getProjectRoot(), "private.pem"))
	if err != nil {
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return "", err
	}

	return token.SignedString(key)
}

// SetTestContext injects userID, clinicaID and userRole into a gin context
func SetTestContext(c *gin.Context, userID uint, clinicaID uint, role string) {
	c.Set("userID", userID)
	c.Set("clinicaID", clinicaID)
	c.Set("userRole", role)
}
