package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        var tokenString string
        
        // 1. Try to get from Authorization Header
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token format is required"})
                return
            }
        } else {
            // 2. Try to get from HttpOnly Cookie
            cookie, err := c.Cookie("auth_token")
            if err != nil {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required (missing header or cookie)"})
                return
            }
            tokenString = cookie
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(jwtSecret), nil
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

            if role, ok := claims["tipoUsuario"].(string); ok {
                c.Set("userRole", role)
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
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: role not permitted"})
            return
        }

        // --- NEW: Granular Clinic Access Verification (Residency Check) ---
        // ADMIN_TOTAL has global access and doesn't need a specific Funcionario/Dentista record
        if user.TipoUsuario == "ADMIN_TOTAL" {
            c.Next()
            return
        }

        clinicaIDRaw, _ := c.Get("clinicaID")
        clinicaID, ok := clinicaIDRaw.(uint)
        if !ok {
             c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
             return
        }

        // Check if user is an active employee or dentist in THIS clinic
        var activeResidency bool
        
        // 1. Check Funcionario status
        var funcCount int64
        db.Model(&models.Funcionario{}).
            Where("usuario_id = ? AND clinica_id = ? AND status = ?", user.ID, clinicaID, "ATIVO").
            Count(&funcCount)
        
        if funcCount > 0 {
            activeResidency = true
        } else {
            // 2. Check Dentista status (if not found as active funcionario)
            var denCount int64
            db.Model(&models.Dentista{}).
                Where("usuario_id = ? AND clinica_id = ?", user.ID, clinicaID).
                Count(&denCount)
            if denCount > 0 {
                activeResidency = true
            }
        }

        if !activeResidency {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "Acesso negado: seu vínculo com esta clínica está inativo ou não existe",
            })
            return
        }

        c.Next()
    }
}
