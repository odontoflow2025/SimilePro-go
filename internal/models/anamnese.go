package models

import (
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
