package models

import (
	"time"

	"gorm.io/gorm"
)

type Dentista struct {
    ID                 uint           `gorm:"primaryKey" json:"id"`
    UsuarioID          uint           `gorm:"uniqueIndex;not null" json:"usuarioId"`
    Usuario            User           `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
    ClinicaID          uint           `json:"clinicaId"`
    Clinica            *Clinica       `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CRO                string         `gorm:"uniqueIndex;not null" json:"cro"`
    Especialidade      string         `json:"especialidade"`
    PorcentagemRepasse float64        `json:"porcentagemRepasse"`
    DataContratacao    time.Time      `json:"dataContratacao"`
    CreatedAt          time.Time      `json:"createdAt"`
    UpdatedAt          time.Time      `json:"updatedAt"`
    DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}
