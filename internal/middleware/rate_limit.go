package middleware

import (
	"context"
	"net/http"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"odonto-flow-go/internal/database"
)

// LoginRateLimiter implementa Token Bucket via Redis limitando a 5 tentativas por IP a cada 5 minutos
func LoginRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:login:%s", ip)
		ctx := context.Background()

		if database.RedisClient == nil {
			c.Next()
			return
		}

		// Incrementa a chave de rate limit
		count, err := database.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			// Em caso de falha no Redis, permite a passagem para evitar downtime
			c.Next()
			return
		}

		// Na primeira requisição, define a janela de 5 minutos
		if count == 1 {
			database.RedisClient.Expire(ctx, key, 5*time.Minute)
		}

		if count > 5 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Muitas tentativas de login falhas. Tente novamente em 5 minutos.",
			})
			return
		}

		c.Next()
	}
}
