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

type NotaFiscalEntrada struct {
	ID          uint      `gorm:"primaryKey"`
	Numero      string    `gorm:"not null"`
	Fornecedor  string    `gorm:"not null"`
	ValorTotal  float64   `gorm:"not null"`
	TransacaoID *uint
	Itens       []ItemNF  `gorm:"foreignKey:NotaFiscalID"`
}

type ItemNF struct {
	ID            uint    `gorm:"primaryKey"`
	NotaFiscalID  uint    `gorm:"not null"`
	ProdutoID     uint    `gorm:"not null"`
	Produto       Produto `gorm:"foreignKey:ProdutoID"`
	Quantidade    float64 `gorm:"not null"`
	PrecoUnitario float64 `gorm:"not null"`
	Subtotal      float64 `gorm:"not null"`
}

type Produto struct {
	ID   uint   `gorm:"primaryKey"`
	Nome string `gorm:"not null"`
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=Simile Pro port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	targetID := 1803

	var nf NotaFiscalEntrada
	err = db.Preload("Itens.Produto").Where("transacao_id = ?", targetID).First(&nf).Error
	if err != nil {
		log.Fatal("NF not found for transaction 1803:", err)
	}

	fmt.Printf("NF Details:\nID: %v | Numero: %s | Fornecedor: %s | Total: %.2f\n", nf.ID, nf.Numero, nf.Fornecedor, nf.ValorTotal)
	fmt.Println("\nItems:")
	for _, item := range nf.Itens {
		fmt.Printf("- %s: %.2f un x R$ %.2f = R$ %.2f\n", item.Produto.Nome, item.Quantidade, item.PrecoUnitario, item.Subtotal)
	}
}
