package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
)

// getClinicaIDFromContext extracts current clinic ID from Gin context securely
func getClinicaIDFromContext(c *gin.Context) (uint, error) {
	clinicaIDRaw, exists := c.Get("clinicaID")
	if !exists {
		return 0, errors.New("clínica não identificada no contexto")
	}

	switch v := clinicaIDRaw.(type) {
	case uint:
		return v, nil
	case float64:
		return uint(v), nil
	default:
		return 0, errors.New("formato de ID da clínica inválido")
	}
}

// getUserRoleFromContext extracts current user role from Gin context securely
func getUserRoleFromContext(c *gin.Context) string {
	userRoleRaw, exists := c.Get("userRole")
	if !exists {
		return ""
	}

	if role, ok := userRoleRaw.(string); ok {
		return role
	}
	return ""
}
