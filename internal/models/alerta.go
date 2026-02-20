package models

import (
	"time"

	"gorm.io/gorm"
)

type NivelAlerta string

const (
	NivelAlertaGravissimo NivelAlerta = "GRAVISSIMO" // Vermelho
	NivelAlertaGrave      NivelAlerta = "GRAVE"      // Laranja
	NivelAlertaAtencao    NivelAlerta = "ATENCAO"    // Amarelo
)

type Alerta struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PacienteID  uint           `gorm:"not null;index" json:"pacienteId"`
	Paciente    Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
	Nivel       NivelAlerta    `gorm:"not null" json:"nivel"`
	Tipo        string         `json:"tipo"` // Diabetes, Alergia, Cardiáco, Outro
	Descricao   string         `gorm:"type:text" json:"descricao"`
	CriadoPorID uint           `json:"criadoPorId"`
	CriadoPor   User           `gorm:"foreignKey:CriadoPorID" json:"criadoPor,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
