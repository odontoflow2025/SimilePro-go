package models

import (
	"SimilePro-go/internal/utils"
	"time"

	"gorm.io/gorm"
)

type Anamnese struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    PacienteID   uint           `gorm:"uniqueIndex;not null" json:"pacienteId"`
    Paciente     Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    HistoricoMedico string      `gorm:"type:text" json:"historicoMedico"` // Doenças pré-existentes, cirurgias
    Alergias        string      `gorm:"type:text" json:"alergias"`
    Medicamentos    string      `gorm:"type:text" json:"medicamentos"` // Em uso contínuo
    Fumante         bool        `json:"fumante"`
    Gestante        bool        `json:"gestante"`
    Observacoes     string      `gorm:"type:text" json:"observacoes"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Anamnese) BeforeSave(tx *gorm.DB) (err error) {
	if a.Observacoes != "" {
		a.Observacoes, err = utils.EncryptAESGCM(a.Observacoes)
		if err != nil {
			return err
		}
	}
	if a.HistoricoMedico != "" {
		a.HistoricoMedico, err = utils.EncryptAESGCM(a.HistoricoMedico)
		if err != nil {
			return err
		}
	}
	if a.Alergias != "" {
		a.Alergias, err = utils.EncryptAESGCM(a.Alergias)
		if err != nil {
			return err
		}
	}
	if a.Medicamentos != "" {
		a.Medicamentos, err = utils.EncryptAESGCM(a.Medicamentos)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Anamnese) AfterFind(tx *gorm.DB) (err error) {
	if a.Observacoes != "" {
		a.Observacoes, err = utils.DecryptAESGCM(a.Observacoes)
		if err != nil {
			return err
		}
	}
	if a.HistoricoMedico != "" {
		a.HistoricoMedico, err = utils.DecryptAESGCM(a.HistoricoMedico)
		if err != nil {
			return err
		}
	}
	if a.Alergias != "" {
		a.Alergias, err = utils.DecryptAESGCM(a.Alergias)
		if err != nil {
			return err
		}
	}
	if a.Medicamentos != "" {
		a.Medicamentos, err = utils.DecryptAESGCM(a.Medicamentos)
		if err != nil {
			return err
		}
	}
	return nil
}

