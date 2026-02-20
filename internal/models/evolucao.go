package models

import (
	"time"

	"gorm.io/gorm"
)

type Evolucao struct {
    ID                     uint           `gorm:"primaryKey" json:"id"`
    PacienteID             uint           `gorm:"not null" json:"pacienteId"`
    Paciente               Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    DentistaID             uint           `gorm:"not null" json:"dentistaId"`
    Dentista               Dentista       `gorm:"foreignKey:DentistaID" json:"dentista,omitempty"`
    ClinicaID              uint           `gorm:"not null" json:"clinicaId"`
    Clinica                Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Descricao              string         `gorm:"type:text" json:"descricao"`
    Data                   time.Time      `json:"data"`
    ProcedimentoRealizadoID *uint          `json:"procedimentoRealizadoId"`
    ProcedimentoRealizado   *Procedimento  `gorm:"foreignKey:ProcedimentoRealizadoID" json:"procedimentoRealizado,omitempty"`
    ItemPlanoID            *uint          `json:"itemPlanoId"` // Link to treatment plan item
    CreatedAt              time.Time      `json:"createdAt"`
    UpdatedAt              time.Time      `json:"updatedAt"`
    DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}
