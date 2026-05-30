package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func sanitizePayload(data interface{}) interface{} {
	blocklist := map[string]bool{
		"senha": true, "password": true, "cpf": true,
		"historicoMedico": true, "alergias": true, "observacoes": true, "medicamentos": true,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, val := range v {
			if blocklist[key] {
				sanitized[key] = "[CENSURADO - PII/PHI]"
			} else {
				sanitized[key] = sanitizePayload(val)
			}
		}
		return sanitized
	case []interface{}:
		sanitizedList := make([]interface{}, len(v))
		for i, val := range v {
			sanitizedList[i] = sanitizePayload(val)
		}
		return sanitizedList
	}
	return data
}

func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	auditService := services.NewAuditService(db)

	return func(c *gin.Context) {
		// Only audit mutating methods
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Read Body safely
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process request
		c.Next()

		// Post-process logic (After handler)
		
		// Only audit successful operations (2xx) or maybe all? Let's generic to all for now.
		// Usually we care if it worked.
		if c.Writer.Status() >= 400 {
			// Optional: Audit failures too? Let's skip failures to reduce noise for now, unless 401/403
			return
		}

		// Extract context info
		userIDVal, _ := c.Get("userID")
		clinicaIDVal, _ := c.Get("clinicaID")
		
		var userID, clinicaID uint

		if userIDVal != nil {
			switch v := userIDVal.(type) {
			case float64: userID = uint(v)
			case uint:    userID = v
			case int:     userID = uint(v)
			}
		}

		if clinicaIDVal != nil {
			switch v := clinicaIDVal.(type) {
			case float64: clinicaID = uint(v)
			case uint:    clinicaID = v
			case int:     clinicaID = uint(v)
			}
		}

		// determine Action
		var action models.AuditAction
		switch c.Request.Method {
		case http.MethodPost:
			action = models.AuditActionCreate
		case http.MethodPut, http.MethodPatch:
			action = models.AuditActionUpdate
		case http.MethodDelete:
			action = models.AuditActionDelete
		default:
			action = "UNKNOWN"
		}

		// Determine Resource from Path
		// /api/pacientes/123 -> Resource: PACIENTES, ID: 123
		pathParts := strings.Split(strings.TrimPrefix(c.Request.URL.Path, "/api/"), "/")
		recurso := ""
		recursoID := ""
		
		if len(pathParts) > 0 {
			recurso = strings.ToUpper(pathParts[0])
		}
		if len(pathParts) > 1 {
			recursoID = pathParts[1]
		}

		// Parse Body for JSON log
		var dados interface{}
		if len(bodyBytes) > 0 {
			var parsedData interface{}
			if err := json.Unmarshal(bodyBytes, &parsedData); err == nil {
				dados = sanitizePayload(parsedData)
			}
		}

		// Log it
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		auditService.Log(userID, clinicaID, action, recurso, recursoID, nil, dados, ip, userAgent)
	}
}
