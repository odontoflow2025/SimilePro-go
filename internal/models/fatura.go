package models

import (
	"time"

	"gorm.io/gorm"
)

type StatusFatura string

const (
    StatusFaturaPendente StatusFatura = "PENDENTE"
    StatusFaturaPaga     StatusFatura = "PAGA"
    StatusFaturaCancelada StatusFatura = "CANCELADA"
)

type Fatura struct {
    ID                uint           `gorm:"primaryKey" json:"id"`
    ClinicaID         uint           `gorm:"not null" json:"clinicaId"`
    Clinica           Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    PacienteID        uint           `gorm:"not null" json:"pacienteId"`
    Paciente          Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    PlanoTratamentoID *uint          `json:"planoTratamentoId"`
    PlanoTratamento   *PlanoTratamento `gorm:"foreignKey:PlanoTratamentoID" json:"planoTratamento,omitempty"`
    ValorTotal        float64        `json:"valorTotal"`
    Status            StatusFatura   `json:"status" gorm:"default:PENDENTE"`
    DataGeracao       time.Time      `json:"dataGeracao"`
    DataVencimento    time.Time      `json:"dataVencimento"`
    CreatedAt         time.Time      `json:"createdAt"`
    UpdatedAt         time.Time      `json:"updatedAt"`
    DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
