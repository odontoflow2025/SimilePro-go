package models

import (
	"time"

	"gorm.io/gorm"
)

type Convenio struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    Nome         string         `gorm:"not null" json:"nome"`
    RegistroANS  string         `json:"registroAns"`
    TabelaPrecos string         `gorm:"type:text" json:"tabelaPrecos"` // JSON string for now
    ClinicaID    uint           `gorm:"not null" json:"clinicaId"`
    Clinica      Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt    time.Time      `json:"createdAt"`
    UpdatedAt    time.Time      `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
