package main

import (
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	email := "jefe.gomes154@gmail.com"
	log.Printf("Atualizando permissão para %s...", email)

	result := db.Model(&models.User{}).Where("email = ?", email).Update("tipo_usuario", "ADMIN_TOTAL")
	
	if result.Error != nil {
		log.Fatalf("Erro ao atualizar: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Println("Usuário não encontrado! Verifique se o email está correto no banco.")
	} else {
		log.Println("Sucesso! Usuário agora é ADMIN_TOTAL.")
	}
}
