package main

import (
	"fmt"
	"log"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Transacao struct {
	ID              uint            `gorm:"primaryKey"`
	ClinicaID       uint            `gorm:"not null"`
	Descricao       string          `gorm:"not null"`
	Valor           float64         `gorm:"not null"`
	Tipo            string          `gorm:"not null"`
	Status          string          `gorm:"default:PENDENTE"`
	DataVencimento  time.Time
	CreatedAt       time.Time
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=Simile Pro port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Today: 2026-03-12
	start := "2026-03-12 00:00:00"
	end := "2026-03-12 23:59:59"

	var transactions []Transacao
	db.Where("data_vencimento BETWEEN ? AND ?", start, end).Order("created_at desc").Find(&transactions)

	fmt.Printf("Found %d transactions due today (2026-03-12):\n", len(transactions))
	totalDespesas := 0.0
	totalReceitas := 0.0

	for _, t := range transactions {
		fmt.Printf("ID: %v | Desc: %s | Valor: %.2f | Tipo: %s | Status: %s | Vencimento: %v\n", 
			t.ID, t.Descricao, t.Valor, t.Tipo, t.Status, t.DataVencimento.Format("2006-01-02 15:04:05"))
		if t.Tipo == "DESPESA" {
			totalDespesas += t.Valor
		} else {
			totalReceitas += t.Valor
		}
	}
	fmt.Printf("\nSummary for 2026-03-12:\nTotal Despesas: %.2f\nTotal Receitas: %.2f\nBalance: %.2f\n", totalDespesas, totalReceitas, totalReceitas - totalDespesas)
}
