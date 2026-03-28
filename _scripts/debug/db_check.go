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
		log.Fatalf("Config error: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("DB Connect error: %v", err)
	}

	var count int64
	db.Model(&models.Transacao{}).Count(&count)
	log.Printf("Total Transacoes: %d", count)

	var receitas int64
	db.Model(&models.Transacao{}).Where("tipo = ?", "RECEITA").Count(&receitas)
	log.Printf("Receitas: %d", receitas)

    var despesas int64
	db.Model(&models.Transacao{}).Where("tipo = ?", "DESPESA").Count(&despesas)
	log.Printf("Despesas: %d", despesas)
}
