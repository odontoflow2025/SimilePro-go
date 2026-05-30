package models

import (
	"time"

	"gorm.io/gorm"
)

type StatusPlano string

const (
    StatusPlanoOrcamento StatusPlano = "ORCAMENTO"
    StatusPlanoAprovado  StatusPlano = "APROVADO"
    StatusPlanoConcluido StatusPlano = "CONCLUIDO"
    StatusPlanoCancelado StatusPlano = "CANCELADO"
)

type PlanoTratamento struct {
    ID          uint              `gorm:"primaryKey" json:"id"`
    PacienteID  uint              `gorm:"not null" json:"pacienteId"`
    Paciente    Paciente          `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    DentistaID  uint              `gorm:"not null" json:"dentistaId"`
    Dentista    Dentista          `gorm:"foreignKey:DentistaID" json:"dentista,omitempty"`
    ClinicaID   uint              `gorm:"not null" json:"clinicaId"`
    Clinica     Clinica           `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Status      StatusPlano       `json:"status" gorm:"default:ORCAMENTO"`
    ValorTotal  float64           `json:"valorTotal"`
    Itens       []ItemPlano       `gorm:"foreignKey:PlanoTratamentoID" json:"itens,omitempty"`
    Version     uint              `gorm:"default:1" json:"-"`
    CreatedAt   time.Time         `json:"createdAt"`
    UpdatedAt   time.Time         `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt    `gorm:"index" json:"-"`
}

type ItemPlano struct {
    ID                uint           `gorm:"primaryKey" json:"id"`
    PlanoTratamentoID uint           `gorm:"not null" json:"planoTratamentoId"`
    ProcedimentoID    uint           `gorm:"not null" json:"procedimentoId"`
    Procedimento      Procedimento   `gorm:"foreignKey:ProcedimentoID" json:"procedimento,omitempty"`
    ValorUnitario     float64        `json:"valorUnitario"`
    Quantidade        int            `json:"quantidade"`
    DenteRegiao       string         `json:"denteRegiao"` // Ex: "18", "Superior"
    Concluido         bool           `json:"concluido" gorm:"default:false"`
    CreatedAt         time.Time      `json:"createdAt"`
    UpdatedAt         time.Time      `json:"updatedAt"`
    DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
