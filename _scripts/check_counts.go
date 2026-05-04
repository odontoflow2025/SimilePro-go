package main

import (
	"fmt"
	"log"
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

	var pacientes int64
	db.Model(&models.Paciente{}).Count(&pacientes)

	var agendamentos int64
	db.Model(&models.Agendamento{}).Count(&agendamentos)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	var pacMes int64
	db.Model(&models.Paciente{}).Where("created_at >= ?", startOfMonth).Count(&pacMes)

	var agMes int64
	db.Model(&models.Agendamento{}).Where("data_hora_inicio >= ?", startOfMonth).Count(&agMes)

	fmt.Printf("Total Pacientes: %d, Pacientes Mes: %d\n", pacientes, pacMes)
	fmt.Printf("Total Agendamentos: %d, Agendamentos Mes: %d\n", agendamentos, agMes)
}
