package middleware

import (
	"strings"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func CORSMiddleware() gin.HandlerFunc {
	// Carrega origens no startup
	allowedStr := viper.GetString("ALLOWED_ORIGINS")
	if allowedStr == "" {
		allowedStr = os.Getenv("ALLOWED_ORIGINS")
	}
	allowedOrigins := strings.Split(allowedStr, ",")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		isAllowed := false
		for _, allowed := range allowedOrigins {
			allowed = strings.TrimSpace(allowed)
			if allowed == origin || allowed == "*" {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Bloqueia implicitamente ao não refletir
			c.Writer.Header().Set("Access-Control-Allow-Origin", "null")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
