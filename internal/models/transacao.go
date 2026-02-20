package models

import (
	"time"

	"gorm.io/gorm"
)

type TipoTransacao string
type StatusTransacao string

const (
    TipoTransacaoReceita TipoTransacao = "RECEITA"
    TipoTransacaoDespesa TipoTransacao = "DESPESA"

    StatusTransacaoPendente StatusTransacao = "PENDENTE"
    StatusTransacaoPago     StatusTransacao = "PAGO"
    StatusTransacaoCancelado StatusTransacao = "CANCELADO"
)

type Transacao struct {
    ID              uint            `gorm:"primaryKey" json:"id"`
    ClinicaID       uint            `gorm:"not null" json:"clinicaId"`
    Clinica         Clinica         `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    PacienteID      *uint           `json:"pacienteId"`
    Paciente        *Paciente       `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    Descricao       string          `gorm:"not null" json:"descricao"`
    Valor           float64         `gorm:"not null" json:"valor"`
    Tipo            TipoTransacao   `gorm:"not null" json:"tipo"`
    Status          StatusTransacao `gorm:"default:PENDENTE" json:"status"`
    DataVencimento  time.Time       `json:"dataVencimento"`
    DataPagamento   *time.Time      `json:"dataPagamento"`
    Categoria       string          `json:"categoria"` // Ex: "Consultas", "Luz", "Aluguel"
    FormaPagamento  string          `json:"formaPagamento"` // "Dinheiro", "Cartao", "Pix"
    CreatedAt       time.Time       `json:"createdAt"`
    UpdatedAt       time.Time       `json:"updatedAt"`
    DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}
