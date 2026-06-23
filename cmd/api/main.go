package main

import (
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/routes"

	"github.com/gin-gonic/gin"
)

// @title           Simile Pro API
// @version         1.0
// @description     Backend API for Simile Pro dental management system.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

	// Security: Fail-fast if encryption key is not strictly 32 bytes (AES-256)
	if len(cfg.EncryptionKey) != 32 {
		log.Fatalf("CRITICAL SECURITY ERROR: ENCRYPTION_KEY must be exactly 32 bytes for AES-256. Found %d bytes. System halted.", len(cfg.EncryptionKey))
	}

    // Connect to database
    db, err := database.Connect(cfg)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Connect to Redis
    database.ConnectRedis(cfg.RedisAddr)

    // Initialize Router
    r := gin.Default()
    // Aceitar IP do container Docker proxy/Cloudflared para RateLimiting e Auditoria
    r.SetTrustedProxies(nil)

    // Setup Routes
    routes.SetupRoutes(r, db, cfg)

    // Run Server
    log.Println("Server running on port " + cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
