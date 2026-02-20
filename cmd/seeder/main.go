package main

import (
	"fmt"
	"log"
	"math/rand"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Load config and connect to DB
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Adjust DB host for local script execution if needed, or assume env is correct
	// If running via `go run`, it uses .env
	
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect via database.Connect: %v", err)
	}

	log.Println("Seeder started...")

	// 1. Find or Create Clinic "Odonto Pró"
	var clinica models.Clinica
	err = db.Where("nome_fantasia ILIKE ?", "%Odonto Pró%").First(&clinica).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("Clinica 'Odonto Pró' not found. Creating it...")
			
			// Needs a Responsible User
			user := models.User{
				Nome:        "Dono Odonto Pró",
				Email:       "dono@odontopro.com",
				CPF:         "00000000000",
				TipoUsuario: "ADMIN_TOTAL",
				SenhaHash:   hashPassword("123456"), 
			}
			if err := db.Create(&user).Error; err != nil {
				log.Fatalf("Failed to create user: %v", err)
			}

			clinica = models.Clinica{
				NomeFantasia:    "Odonto Pró",
				RazaoSocial:     "Odonto Pró Ltda",
				CNPJ:            "00.000.000/0001-00",
				ResponsavelID:   user.ID,
				Plano:           "PRO",
				MaxFuncionarios: 9999,
			}
			if err := db.Create(&clinica).Error; err != nil {
				log.Fatalf("Failed to create clinica: %v", err)
			}
		} else {
			log.Fatalf("Failed to query clinica: %v", err)
		}
	} else {
		log.Printf("Found Clinica: %s (ID: %d)", clinica.NomeFantasia, clinica.ID)
		// Update plan just in case
		db.Model(&clinica).Updates(map[string]interface{}{
			"plano": "PRO", 
			"max_funcionarios": 9999,
		})
	}

	// 2. Find or Create Dentist
	var dentista models.Dentista
	if err := db.Where("id IN (SELECT dentista_id FROM agendamentos WHERE clinica_id = ?)", clinica.ID).First(&dentista).Error; err != nil {
		// Create new dentist
		userDentist := models.User{
			Nome:        "Dr. Exemplo",
			Email:       "dr.exemplo@odontopro.com",
			CPF:         "11111111111",
			TipoUsuario: "DENTISTA",
			SenhaHash:   hashPassword("123456"),
		}
		db.Where("email = ?", userDentist.Email).FirstOrCreate(&userDentist)

		dentista = models.Dentista{
			UsuarioID:     userDentist.ID,
			CRO:           "SP-12345",
			Especialidade: "Geral",
		}
		db.Where("usuario_id = ?", userDentist.ID).FirstOrCreate(&dentista)
	}
	log.Printf("Using Dentist ID: %d", dentista.ID)

	// 3. Ensure Patients
	var pacientes []models.Paciente
	if err := db.Where("clinica_id = ?", clinica.ID).Find(&pacientes).Error; err != nil || len(pacientes) < 10 {
		log.Println("Creating random patients...")
		for i := 0; i < 20; i++ {
			p := models.Paciente{
				Nome:             fmt.Sprintf("Paciente Exemplo %d", i+1),
				CPF:              fmt.Sprintf("000%03d%03d00", i, i),
				ClinicaID:        clinica.ID,
				TelefonePrincipal: "11999999999",
			}
			if err := db.Create(&p).Error; err == nil {
				pacientes = append(pacientes, p)
			}
		}
	}

	// 4. Loop Dates (2020-01-03 to 2026-02-14)
	startDate := time.Date(2020, 1, 3, 0, 0, 0, 0, time.Local)
	endDate := time.Date(2026, 2, 14, 23, 59, 59, 0, time.Local)

	type Shift struct {
		StartHour int
		StartMin  int
		EndHour   int
		EndMin    int
	}

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		if weekday == time.Sunday {
			continue // Closed
		}

		var shifts []Shift
		if weekday == time.Saturday {
			// Sat: 08:00 - 13:30
			shifts = []Shift{{8, 0, 13, 30}}
		} else {
			// Mon-Fri: 08:00 - 11:30 and 13:00 - 21:00
			shifts = []Shift{
				{8, 0, 11, 30},
				{13, 0, 21, 0},
			}
		}

		// Generate random appointments for today
		numAppts := rand.Intn(6) + 2 // 2 to 7 appointments per day

		for i := 0; i < numAppts; i++ {
			// Pick random patient
			patient := pacientes[rand.Intn(len(pacientes))]
			
			// Pick random shift
			shift := shifts[rand.Intn(len(shifts))]

			// Random start time within shift (simple logic)
			// Convert shift to minutes from midnight
			startMinTotal := shift.StartHour*60 + shift.StartMin
			endMinTotal := shift.EndHour*60 + shift.EndMin
			duration := 30 // 30 min appointment
			
			// Random start minute
			maxStart := endMinTotal - duration
			if maxStart <= startMinTotal { continue }
			
			randomMin := rand.Intn(maxStart-startMinTotal) + startMinTotal
			
			apptHour := randomMin / 60
			apptMin := randomMin % 60

			apptTime := time.Date(d.Year(), d.Month(), d.Day(), apptHour, apptMin, 0, 0, time.Local)
			apptEndTime := apptTime.Add(time.Minute * time.Duration(duration))

			// Create Appointment
			appt := models.Agendamento{
				PacienteID:     patient.ID,
				DentistaID:     dentista.ID,
				ClinicaID:      clinica.ID,
				DataHoraInicio: apptTime,
				DataHoraFim:    apptEndTime,
				Status:         models.StatusAgendamentoAtendido, // Completed to count as revenue
				Motivo:         "Consulta de Rotina / Tratamento",
			}

			if err := db.Create(&appt).Error; err != nil {
				log.Printf("Error creating appointment: %v", err)
				continue
			}

			// Create Transaction (Revenue)
			// Random Value between 100 and 500
			val := float64(rand.Intn(400) + 100)
			
			trans := models.Transacao{
				ClinicaID:      clinica.ID,
				PacienteID:     &patient.ID,
				Descricao:      fmt.Sprintf("Atendimento ref. Agendamento #%d", appt.ID),
				Valor:          val,
				Tipo:           models.TipoTransacaoReceita,
				Status:         models.StatusTransacaoPago,
				DataVencimento: apptTime,
				DataPagamento:  &apptTime,
				Categoria:      "Tratamentos",
				FormaPagamento: "Cartao",
			}
			db.Create(&trans)
		}
		
		// Optional: Add some expenses (Despesas) occasionally
		if rand.Intn(10) == 0 { // 1 in 10 days
			expense := models.Transacao{
				ClinicaID:      clinica.ID,
				Descricao:      "Compra de Materiais / Contas",
				Valor:          float64(rand.Intn(1000) + 200),
				Tipo:           models.TipoTransacaoDespesa,
				Status:         models.StatusTransacaoPago,
				DataVencimento: d,
				DataPagamento:  &d,
				Categoria:      "Operacional",
				FormaPagamento: "Boleto",
			}
			db.Create(&expense)
		}
	}

	log.Println("Seeding completed successfully!")
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}
