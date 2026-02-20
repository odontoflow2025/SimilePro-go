package models

import (
	"time"

	"gorm.io/gorm"
)

type AcessoProntuario struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	PacienteID      uint           `gorm:"not null;index" json:"pacienteId"`
	Paciente        Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
	UsuarioID       uint           `gorm:"not null;index" json:"usuarioId"`
	Usuario         User           `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	NumeroProtocolo string         `gorm:"not null;index" json:"numeroProtocolo"`
	Justificativa   string         `gorm:"type:text" json:"justificativa"`
	Expiracao       time.Time      `json:"expiracao"`
	CreatedAt       time.Time      `json:"createdAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
