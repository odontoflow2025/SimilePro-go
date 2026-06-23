package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"odonto-flow-go/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=root dbname=Simile Pro port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var clinica models.Clinica
	if err := db.First(&clinica).Error; err != nil {
		log.Fatal("No clinic found. Please seed a clinic first.")
	}

	var paciente models.Paciente
	db.First(&paciente) // Patient is optional but good to have

	startDate := time.Date(2026, 2, 9, 10, 0, 0, 0, time.Local)
	endDate := time.Date(2026, 4, 28, 18, 0, 0, 0, time.Local)

	procedimentos := []string{"Limpeza (Profilaxia)", "Tratamento de Canal", "Clareamento Dental", "Extração de Siso", "Restauração em Resina", "Manutenção de Aparelho"}
	despesas := []string{"Material de Consumo (Luvas, Máscaras)", "Conta de Luz", "Internet", "Laboratório de Prótese", "Produtos de Limpeza"}

	rand.Seed(time.Now().UnixNano())

	var transacoes []models.Transacao

	// Generate 2-4 transactions per day
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Skip Sundays
		if d.Weekday() == time.Sunday {
			continue
		}

		// 1 to 3 Receitas (Entradas)
		numReceitas := rand.Intn(3) + 1
		for i := 0; i < numReceitas; i++ {
			procName := procedimentos[rand.Intn(len(procedimentos))]
			valor := float64(rand.Intn(400) + 100) // 100 to 500
			
			var pacID *uint
			if paciente.ID != 0 {
				pacID = &paciente.ID
			}

			t := models.Transacao{
				ClinicaID:      clinica.ID,
				PacienteID:     pacID,
				Descricao:      "Procedimento: " + procName,
				Valor:          valor,
				Tipo:           models.TipoTransacaoReceita,
				Status:         models.StatusTransacaoPago,
				DataVencimento: d,
				DataPagamento:  &d,
				Categoria:      "Procedimentos",
				FormaPagamento: "Pix",
				CreatedAt:      d,
				UpdatedAt:      d,
			}
			transacoes = append(transacoes, t)
		}

		// 1 Despesa (Saída) randomly (50% chance per day)
		if rand.Float32() > 0.5 {
			despName := despesas[rand.Intn(len(despesas))]
			valor := float64(rand.Intn(200) + 50) // 50 to 250

			t := models.Transacao{
				ClinicaID:      clinica.ID,
				Descricao:      despName,
				Valor:          valor,
				Tipo:           models.TipoTransacaoDespesa,
				Status:         models.StatusTransacaoPago,
				DataVencimento: d,
				DataPagamento:  &d,
				Categoria:      "Custos Fixos/Variáveis",
				FormaPagamento: "Boleto",
				CreatedAt:      d,
				UpdatedAt:      d,
			}
			transacoes = append(transacoes, t)
		}
	}

	result := db.Create(&transacoes)
	if result.Error != nil {
		log.Fatalf("Error inserting transactions: %v", result.Error)
	}

	fmt.Printf("Successfully inserted %d transactions (entradas e saídas) from Feb 9 to Apr 28!\n", len(transacoes))
}
