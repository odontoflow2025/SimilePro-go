package models

import (
	"time"

	"gorm.io/gorm"
)

type Funcionario struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    UsuarioID    uint           `gorm:"uniqueIndex;not null" json:"usuarioId"`
    Usuario      User           `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
    ClinicaID    uint           `gorm:"not null" json:"clinicaId"`
    Clinica      Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Cargo        string         `json:"cargo"` // RECEPCIONISTA, ASB, TSB, FINANCEIRO, etc.
    DataAdmissao time.Time      `json:"dataAdmissao"`
    Salario      float64        `json:"salario"`
    CreatedAt    time.Time      `json:"createdAt"`
    UpdatedAt    time.Time      `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
