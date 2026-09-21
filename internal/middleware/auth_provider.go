package middleware

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"SimilePro-go/internal/database"
	"context"
)


type AuthProvider struct {
	publicKey *rsa.PublicKey
}

func NewAuthProvider(publicKeyPath string) *AuthProvider {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatalf("Failed to read public key from %s: %v", publicKeyPath, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}

	return &AuthProvider{
		publicKey: publicKey,
	}
}

func (a *AuthProvider) AuthMiddleware() gin.HandlerFunc {
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
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Retorna a chave pública da memória
			return a.publicKey, nil
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

			if jti, ok := claims["jti"].(string); ok {
				// Verifica a Deny List no Redis
				ctx := context.Background()
				isBlacklisted := int64(0)
				if database.RedisClient != nil {
					isBlacklisted, _ = database.RedisClient.Exists(ctx, "blacklist:"+jti).Result()
				}
				if isBlacklisted > 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sessão revogada"})
					return
				}
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
