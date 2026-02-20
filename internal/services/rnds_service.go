package services

import (
	"errors"
	"fmt"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/rnds"
	"time"

	"gorm.io/gorm"
)

type RndsService struct {
	DB *gorm.DB
}

func NewRndsService(db *gorm.DB) *RndsService {
	return &RndsService{DB: db}
}

// CheckStatus verifies if the clinic has valid RNDS credentials
func (s *RndsService) CheckStatus(clinicaID uint) (bool, string, error) {
	var clinica models.Clinica
	if err := s.DB.First(&clinica, clinicaID).Error; err != nil {
		return false, "", err
	}

	if !clinica.RndsAtivo {
		return false, "RNDS não está ativo para esta clínica", nil
	}

	if clinica.RndsCertificado == "" {
		return false, "Certificado digital não configurado", nil
	}

	// Mock check with Gov API
	// In real world, we would load the .pfx and ping the server
	return true, "Conectado e Operacional", nil
}

// SendPatient sends a patient to RNDS
func (s *RndsService) SendPatient(patientID uint) (string, error) {
	var paciente models.Paciente
	if err := s.DB.Preload("Clinica").First(&paciente, patientID).Error; err != nil {
		return "", err
	}

	// 1. Build FHIR
	fhirPatient, err := rnds.BuildPatientFHIR(paciente)
	if err != nil {
		return "", fmt.Errorf("erro ao construir FHIR: %v", err)
	}

	// 2. Validate Clinic Creds
	if !paciente.Clinica.RndsAtivo {
		return "", errors.New("RNDS desativado")
	}

	// 3. Send (Mock)
	// client := rnds.NewClient(...)
	// resp, err := client.Post("Patient", fhirPatient)
    
    // Silence unused variable linter for now
    _ = fhirPatient
	
	// Mock success response ID from RNDS
	rndsID := fmt.Sprintf("RNDS-%d-%s", patientID, time.Now().Format("20060102"))
	
	return rndsID, nil
}

// SendAttendance sends an encounter
func (s *RndsService) SendAttendance(agendamentoID uint) (string, error) {
    // Similar logic: Fetch Agendamento, Build Encounter FHIR, Send.
    return "RNDS-ATT-MOCK", nil
}
