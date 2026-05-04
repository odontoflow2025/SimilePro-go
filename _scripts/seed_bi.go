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
	dsn := "host=localhost user=postgres password=root dbname=OdontoFlow port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var clinica models.Clinica
	if err := db.First(&clinica).Error; err != nil {
		log.Fatal("No clinic found.")
	}
	
	var dentista models.Dentista
	if err := db.First(&dentista).Error; err != nil {
		log.Fatal("No dentista found.")
	}

	now := time.Now()
	// Let's create some patients and appointments in the current month to show data in the BI indicators.
	rand.Seed(time.Now().UnixNano())

	// Create 15 New Patients this month
	var pacientes []models.Paciente
	for i := 0; i < 15; i++ {
		// Random day this month up to today
		day := rand.Intn(now.Day()) + 1
		creationDate := time.Date(now.Year(), now.Month(), day, 10, 0, 0, 0, time.Local)
		
		p := models.Paciente{
			Nome:        fmt.Sprintf("Novo Paciente Abril %d", i+1),
			Email:       fmt.Sprintf("novoabril%d@exemplo.com", i+1),
			CPF:         fmt.Sprintf("%011d", rand.Int63n(99999999999)),
			CodigoUnico: fmt.Sprintf("PAC-%d", rand.Int63n(9999999)),
			ClinicaID: clinica.ID,
			CreatedAt: creationDate,
			UpdatedAt: creationDate,
		}
		pacientes = append(pacientes, p)
	}
	if err := db.Create(&pacientes).Error; err != nil {
		log.Fatalf("Failed to insert patients: %v", err)
	}
	fmt.Printf("Inserted %d new patients for the current month.\n", len(pacientes))

	// Create Appointments (Agendamentos) for this month to generate Occupancy Rate
	// Let's create about 80 appointments (which is 80 hours out of ~176 capacity = ~45% occupancy)
	var agendamentos []models.Agendamento
	for i := 0; i < 80; i++ {
		day := rand.Intn(28) + 1
		hour := rand.Intn(8) + 9 // 9 AM to 4 PM
		
		startTime := time.Date(now.Year(), now.Month(), day, hour, 0, 0, 0, time.Local)
		endTime := startTime.Add(1 * time.Hour)
		
		// Pick a random patient from the newly created ones
		pac := pacientes[rand.Intn(len(pacientes))]

		ag := models.Agendamento{
			PacienteID:     pac.ID,
			DentistaID:     dentista.ID,
			ClinicaID:      clinica.ID,
			DataHoraInicio: startTime,
			DataHoraFim:    endTime,
			Motivo:         "Consulta de Rotina",
			Status:         models.StatusAgendamentoAgendado,
			CreatedAt:      startTime,
			UpdatedAt:      startTime,
		}
		agendamentos = append(agendamentos, ag)
	}
	if err := db.Create(&agendamentos).Error; err != nil {
		log.Fatalf("Failed to insert appointments: %v", err)
	}
	fmt.Printf("Inserted %d appointments for the current month.\n", len(agendamentos))
}
