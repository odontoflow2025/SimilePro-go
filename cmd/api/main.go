package main

import (
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/routes"

	"github.com/gin-gonic/gin"
)

// @title           OdontoFlow API
// @version         1.0
// @description     Backend API for OdontoFlow dental management system.
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

    // Connect to database
    db, err := database.Connect(cfg)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize Router
    r := gin.Default()

    // Setup Routes
    routes.SetupRoutes(r, db, cfg)

    // Run Server
    log.Println("Server running on port " + cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
