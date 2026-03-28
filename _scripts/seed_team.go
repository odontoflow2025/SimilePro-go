package main

import (
	"fmt"
	"log"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"

	"golang.org/x/crypto/bcrypt"
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

	// Fetch clinic
	var clinica models.Clinica
	if err := db.First(&clinica).Error; err != nil {
		log.Fatalf("No clinic found! Ensure you have at least one clinic. %v", err)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	pass := string(hashed)

	fmt.Println("Seeding team members...")

	cpfBase := 10000000000

	// 5 Dentistas
	for i := 1; i <= 5; i++ {
		cpfVal := cpfBase + i
		u := models.User{
			Nome:        fmt.Sprintf("Dr. Dentista Exemplo %d", i),
			Email:       fmt.Sprintf("dentista%d@clinica.com", i),
			SenhaHash:   pass,
			Telefone:    fmt.Sprintf("1199999000%d", i),
			CPF:         fmt.Sprintf("%d", cpfVal),
			TipoUsuario: "DENTISTA",
			ClinicaID:   clinica.ID,
		}
		if err := db.Create(&u).Error; err == nil {
			d := models.Dentista{
				UsuarioID:     u.ID,
				ClinicaID:     clinica.ID,
				CRO:           fmt.Sprintf("CRO-SP %d000%d", i, i),
				Especialidade: "Clínico Geral",
			}
			db.Create(&d)
			fmt.Printf("Created Dentista: %s\n", u.Nome)
		} else {
             fmt.Printf("Error creating %s: %v\n", u.Nome, err)
        }
	}

	// 5 Tecnicos (Assistentes)
	for i := 1; i <= 5; i++ {
        cpfVal := cpfBase + 10 + i
		u := models.User{
			Nome:        fmt.Sprintf("Técnico ASB %d", i),
			Email:       fmt.Sprintf("tecnico%d@clinica.com", i),
			SenhaHash:   pass,
			Telefone:    fmt.Sprintf("1198888000%d", i),
            CPF:         fmt.Sprintf("%d", cpfVal),
			TipoUsuario: "ASSISTENTE",
			ClinicaID:   clinica.ID,
		}
		if err := db.Create(&u).Error; err == nil {
			f := models.Funcionario{
				UsuarioID: u.ID,
				ClinicaID: clinica.ID,
				Cargo:     "Técnico Especialista",
				Salario:   2500.0,
			}
			db.Create(&f)
			fmt.Printf("Created Técnico: %s\n", u.Nome)
		} else {
             fmt.Printf("Error creating %s: %v\n", u.Nome, err)
        }
	}

	// 2 Atendentes (Recepção)
	for i := 1; i <= 2; i++ {
        cpfVal := cpfBase + 20 + i
		u := models.User{
			Nome:        fmt.Sprintf("Atendente Recepção %d", i),
			Email:       fmt.Sprintf("atendimento%d@clinica.com", i),
			SenhaHash:   pass,
			Telefone:    fmt.Sprintf("1197777000%d", i),
            CPF:         fmt.Sprintf("%d", cpfVal),
			TipoUsuario: "RECEPCIONISTA",
			ClinicaID:   clinica.ID,
		}
		if err := db.Create(&u).Error; err == nil {
			f := models.Funcionario{
				UsuarioID: u.ID,
				ClinicaID: clinica.ID,
				Cargo:     "Atendente / Secretária",
				Salario:   2000.0,
			}
			db.Create(&f)
			fmt.Printf("Created Atendente: %s\n", u.Nome)
		} else {
             fmt.Printf("Error creating %s: %v\n", u.Nome, err)
        }
	}

	// 2 Secretárias (Financeiro/RH)
	for i := 1; i <= 2; i++ {
        cpfVal := cpfBase + 30 + i
		u := models.User{
			Nome:        fmt.Sprintf("Secretária Financeira %d", i),
			Email:       fmt.Sprintf("financeiro%d@clinica.com", i),
			SenhaHash:   pass,
			Telefone:    fmt.Sprintf("1196666000%d", i),
            CPF:         fmt.Sprintf("%d", cpfVal),
			TipoUsuario: "FATURISTA", // or ADMIN_GERENCIAL
			ClinicaID:   clinica.ID,
		}
		if err := db.Create(&u).Error; err == nil {
			f := models.Funcionario{
				UsuarioID: u.ID,
				ClinicaID: clinica.ID,
				Cargo:     "RH e Contabilidade",
				Salario:   3500.0,
			}
			db.Create(&f)
			fmt.Printf("Created Secretária Financeira: %s\n", u.Nome)
		} else {
             fmt.Printf("Error creating %s: %v\n", u.Nome, err)
        }
	}

	fmt.Println("Seeding complete!")
}
