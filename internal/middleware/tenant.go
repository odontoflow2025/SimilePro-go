package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantDBMiddleware injeta uma instância do DB restrita à clínica atual
func TenantDBMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pula a restrição se for um ADMIN_TOTAL (acesso global)
		userRole, _ := c.Get("userRole")
		if userRole == "ADMIN_TOTAL" {
			c.Set("tenantDB", db)
			c.Next()
			return
		}

		// Recupera o clinicaID setado pelo AuthMiddleware
		clinicaIDRaw, exists := c.Get("clinicaID")
		if exists {
			clinicaID, ok := clinicaIDRaw.(uint)
			if ok {
				// Cria um clone do DB com a cláusula WHERE fixa
				tenantDB := db.Where("clinica_id = ?", clinicaID)
				c.Set("tenantDB", tenantDB)
			} else {
				c.Set("tenantDB", db)
			}
		} else {
			// Fallback caso seja rota pública
			c.Set("tenantDB", db)
		}
		c.Next()
	}
}
