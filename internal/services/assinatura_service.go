package services

import (
	"errors"
	"odonto-flow-go/internal/models"

	"gorm.io/gorm"
)

type AssinaturaService struct {
	DB *gorm.DB
}

func NewAssinaturaService(db *gorm.DB) *AssinaturaService {
	return &AssinaturaService{DB: db}
}

// CheckLimit verifica se a clínica pode usar o recurso
// Recursos: "USUARIOS", "DASHBOARD", "FINANCEIRO"
func (s *AssinaturaService) CheckLimit(clinicaID uint, resource string) (bool, error) {
	var clinica models.Clinica
	if err := s.DB.Preload("Dentistas").Preload("Funcionarios").First(&clinica, clinicaID).Error; err != nil {
		return false, err
	}

	// Regras
	switch clinica.Plano {
	case "FREE":
		if resource == "USUARIOS" {
			// Count dentistas
            var dentistas int64
            s.DB.Model(&models.Dentista{}).Where("clinica_id = ?", clinicaID).Count(&dentistas)
            
            // Count funcionarios
            var funcionarios int64
            s.DB.Model(&models.Funcionario{}).Where("clinica_id = ?", clinicaID).Count(&funcionarios)

			total := int(dentistas + funcionarios)
			if clinica.MaxFuncionarios > 0 && total >= clinica.MaxFuncionarios {
				return false, nil
			}
		}
		if resource == "DASHBOARD" || resource == "FINANCEIRO" {
			return false, nil
		}
	case "BASE":
		return true, nil
	case "EDUCACIONAL":
		// Educacional usually has restrictions too?
		if resource == "FINANCEIRO" {
			return false, nil // Assume school doesn't use real caching flow
		}
		return true, nil
	case "PREMIUM":
		return true, nil
	}

	return true, nil
}

func (s *AssinaturaService) UpdatePlan(clinicaID uint, plano string) error {
	var maxFunc int
	switch plano {
	case "FREE":
		maxFunc = 2
	case "BASE":
		maxFunc = 5
	case "PREMIUM":
		maxFunc = 9999
	case "EDUCACIONAL":
		maxFunc = 9999
	default:
		return errors.New("plano inválido")
	}

	return s.DB.Model(&models.Clinica{}).Where("id = ?", clinicaID).Updates(map[string]interface{}{
		"plano":            plano,
		"max_funcionarios": maxFunc,
	}).Error
}
