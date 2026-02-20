package models

import (
	"time"

	"gorm.io/gorm"
)

type Procedimento struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    Nome            string         `gorm:"not null" json:"nome"`
    Codigo          string         `json:"codigo"`
    ValorReferencia float64        `json:"valorReferencia"`
    Ativo           bool           `gorm:"default:true" json:"ativo"`
    ClinicaID       uint           `gorm:"not null" json:"clinicaId"`
    Clinica         Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt       time.Time      `json:"createdAt"`
    UpdatedAt       time.Time      `json:"updatedAt"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
