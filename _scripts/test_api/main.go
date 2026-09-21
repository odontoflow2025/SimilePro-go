package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"SimilePro-go/internal/config"
	"SimilePro-go/internal/database"
	"SimilePro-go/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	var user models.User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("No users found: %v", err)
	}

	fmt.Printf("Testing request for user: %s (ClinicaID: %d)\n", user.Email, user.ClinicaID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         float64(user.ID),
		"name":        user.Nome,
		"email":       user.Email,
		"tipoUsuario": user.TipoUsuario,
		"clinicaId":   float64(user.ClinicaID),
		"exp":         time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}

	req, _ := http.NewRequest("GET", "http://localhost:8080/api/estoque/notas-fiscais", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("HTTP Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", string(body))
}
