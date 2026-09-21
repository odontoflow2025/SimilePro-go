package services

import (
	"errors"
	"SimilePro-go/internal/models"
	"time"

	"gorm.io/gorm"
)

type AssinaturaService struct {
	DB *gorm.DB
}

func NewAssinaturaService(db *gorm.DB) *AssinaturaService {
	return &AssinaturaService{DB: db}
}

// IsActive checks if the clinic's plan is still valid
func (s *AssinaturaService) IsActive(clinica *models.Clinica) bool {
    if clinica.Plano == "FREE" || clinica.Plano == "EDUCACIONAL" {
        return true
    }
    if clinica.PlanoExpiracao == nil {
        return false
    }
    return time.Now().Before(*clinica.PlanoExpiracao) && clinica.PlanoStatus == "ACTIVE"
}

// CheckLimit verifica se a clínica pode usar o recurso
// Recursos: "USUARIOS", "DASHBOARD", "FINANCEIRO", "ESTOQUE", "RELATORIOS"
func (s *AssinaturaService) CheckLimit(clinicaID uint, resource string) (bool, error) {
	var clinica models.Clinica
	if err := s.DB.First(&clinica, clinicaID).Error; err != nil {
		return false, err
	}

    // Check expiration: If expired, treat as FREE
    isActive := s.IsActive(&clinica)
    currentPlan := clinica.Plano
    if !isActive {
        currentPlan = "FREE"
    }

	// Regras
	switch currentPlan {
	case "FREE":
		if resource == "USUARIOS" {
            var dentistas int64
            s.DB.Model(&models.Dentista{}).Where("clinica_id = ?", clinicaID).Count(&dentistas)
            var funcionarios int64
            s.DB.Model(&models.Funcionario{}).Where("clinica_id = ?", clinicaID).Count(&funcionarios)

			total := int(dentistas + funcionarios)
			if total >= clinica.MaxFuncionarios {
				return false, nil
			}
            return true, nil
		}
		// FREE plans don't have these modules
		if resource == "DASHBOARD" || resource == "FINANCEIRO" || resource == "ESTOQUE" || resource == "RELATORIOS" {
			return false, nil
		}
	case "BASE":
        // Base has everything except Multi-Units and advanced BI
        if resource == "RELATORIOS_AVANCADOS" {
            return false, nil
        }
		return true, nil
	case "EDUCACIONAL":
		if resource == "FINANCEIRO" {
			return false, nil
		}
		return true, nil
	case "PREMIUM":
		return true, nil
	}

	return true, nil
}

func (s *AssinaturaService) UpdatePlan(clinicaID uint, plano string, periodicidade string) error {
	var maxFunc int
	var duration time.Duration
    
	switch plano {
	case "FREE":
		maxFunc = 2
	case "BASE":
		maxFunc = 5
        duration = 30 * 24 * time.Hour // Default 30 days
	case "PREMIUM":
		maxFunc = 9999
        duration = 30 * 24 * time.Hour
	case "EDUCACIONAL":
		maxFunc = 9999
	default:
		return errors.New("plano inválido")
	}

    // Specific periodicity for paid plans
    if plano == "BASE" || plano == "PREMIUM" {
        if periodicidade == "ANUAL" {
            duration = 365 * 24 * time.Hour
        } else if periodicidade == "MENSAL" {
            duration = 30 * 24 * time.Hour
        }
    }

    now := time.Now()
    expiracao := now.Add(duration)
    
	return s.DB.Model(&models.Clinica{}).Where("id = ?", clinicaID).Updates(map[string]interface{}{
		"plano":                  plano,
		"max_funcionarios":       maxFunc,
        "plano_expiracao":        expiracao,
        "plano_status":           "ACTIVE",
        "plano_ultimo_pagamento": now,
	}).Error
}
