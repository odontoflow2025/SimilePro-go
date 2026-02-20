package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token format is required"})
            return
        }

        cfg, _ := config.LoadConfig()
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(cfg.JWTSecret), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            if sub, ok := claims["sub"].(float64); ok {
                c.Set("userID", uint(sub))
            } else {
                 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
                 return
            }

            if cid, ok := claims["clinicaId"].(float64); ok {
                c.Set("clinicaID", uint(cid))
            }
        }

        c.Next()
    }
}

func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }

        var user models.User
        if err := db.First(&user, userID).Error; err != nil {
             c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
             return
        }

        roleAllowed := false
        for _, role := range allowedRoles {
            if user.TipoUsuario == role {
                roleAllowed = true
                break
            }
        }

        if !roleAllowed {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            return
        }

        c.Next()
    }
}
