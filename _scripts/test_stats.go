package main

import (
	"fmt"
	"log"
	"time"

	"SimilePro-go/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=root dbname=Simile Pro port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var pacientes int64
	db.Model(&models.Paciente{}).Count(&pacientes)

	var dentistas int64
	db.Model(&models.Dentista{}).Count(&dentistas)

	var consultas int64
	db.Model(&models.Agendamento{}).Count(&consultas) // Total appointments

	// Calculate New Patients this month
	var newPatients int64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	db.Model(&models.Paciente{}).Where("created_at >= ?", startOfMonth).Count(&newPatients)

	// Calculate Occupancy Rate
	var consultasMes int64
	db.Model(&models.Agendamento{}).Where("data_hora_inicio >= ?", startOfMonth).Count(&consultasMes)
	
	var occupancyRate int = 0
	if dentistas > 0 {
		capacidadeMensal := dentistas * 176
		occupancyRate = int((float64(consultasMes) / float64(capacidadeMensal)) * 100)
		if occupancyRate > 100 {
			occupancyRate = 100
		}
	} else if consultasMes > 0 {
	    occupancyRate = 100
	}

	fmt.Printf("Total Pacientes: %d\n", pacientes)
	fmt.Printf("Total Dentistas: %d\n", dentistas)
	fmt.Printf("Total Consultas: %d\n", consultas)
	fmt.Printf("Consultas Mes: %d\n", consultasMes)
	fmt.Printf("New Patients: %d\n", newPatients)
	fmt.Printf("Occupancy Rate: %d\n", occupancyRate)
}
