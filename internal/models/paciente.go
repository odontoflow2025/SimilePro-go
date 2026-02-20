package models

import (
	"time"

	"gorm.io/gorm"
)

type Paciente struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    Nome             string         `gorm:"not null" json:"nome"`
    CPF              string         `gorm:"uniqueIndex" json:"cpf"`
    CodigoUnico      string         `gorm:"uniqueIndex" json:"codigoUnico"`
    RG               string         `json:"rg"`
    DataNascimento   time.Time      `json:"dataNascimento"`
    Genero           string         `json:"genero"` // M, F, O
    TelefonePrincipal string        `json:"telefonePrincipal"`
    Email            string         `json:"email"`
    EnderecoCompleto string         `json:"enderecoCompleto"`
    ClinicaID        uint           `gorm:"not null" json:"clinicaId"`
    Clinica          Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt        time.Time      `json:"createdAt"`
    UpdatedAt        time.Time      `json:"updatedAt"`
    DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
