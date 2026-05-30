package models

import (
	"time"

	"gorm.io/gorm"
)

type CentroCusto struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Nome      string         `gorm:"not null" json:"nome"`
    Descricao string         `json:"descricao"`
    Codigo    string         `json:"codigo"`
    Ativo     bool           `gorm:"default:true" json:"ativo"`
    ClinicaID uint           `gorm:"not null" json:"clinicaId"`
    Clinica   Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TipoConta string

const (
    TipoContaReceita TipoConta = "RECEITA"
    TipoContaDespesa TipoConta = "DESPESA"
    TipoContaAtivo   TipoConta = "ATIVO"
    TipoContaPassivo TipoConta = "PASSIVO"
)

type PlanoConta struct {
    ID             uint           `gorm:"primaryKey" json:"id"`
    Nome           string         `gorm:"not null" json:"nome"`
    Codigo         string         `gorm:"not null" json:"codigo"`
    Tipo           TipoConta      `gorm:"not null" json:"tipo"`
    ContaPaiID     *uint          `json:"contaPaiId"`
    ContaPai       *PlanoConta    `gorm:"foreignKey:ContaPaiID" json:"contaPai,omitempty"`
    SubContas      []PlanoConta   `gorm:"foreignKey:ContaPaiID" json:"subContas,omitempty"`
    AceitaLancamento bool           `gorm:"default:true" json:"aceitaLancamento"`
    ClinicaID      uint           `gorm:"not null" json:"clinicaId"`
    Clinica        Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt      time.Time      `json:"createdAt"`
    UpdatedAt      time.Time      `json:"updatedAt"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// Lançamento Contábil baseado em Partidas Dobradas (Débito/Crédito) p/ DRE e Balanço
type LancamentoContabil struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ClinicaID     uint           `gorm:"not null" json:"clinicaId"`
	Clinica       Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Data          time.Time      `gorm:"not null" json:"data"`
	ContaDebitoID uint           `gorm:"not null" json:"contaDebitoId"`
	ContaDebito   PlanoConta     `gorm:"foreignKey:ContaDebitoID" json:"contaDebito,omitempty"`
	ContaCreditoID uint          `gorm:"not null" json:"contaCreditoId"`
	ContaCredito  PlanoConta     `gorm:"foreignKey:ContaCreditoID" json:"contaCredito,omitempty"`
	Valor         float64        `gorm:"not null" json:"valor"`
	Historico     string         `gorm:"not null" json:"historico"` // Ex: "Recebimento Fatura 123", "Pgto Luz"
	TransacaoID   *uint          `json:"transacaoId"` // Opcional: ligação caso o lançamento venha de uma transação do financeiro operacional
	Transacao     *Transacao     `gorm:"foreignKey:TransacaoID" json:"transacao,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BalancoFinanceiroTenant registra históricos consolidados de tributos e provisões
type BalancoFinanceiroTenant struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ClinicaID    uint      `gorm:"not null;index" json:"clinicaId"`
	Mes          int       `gorm:"not null" json:"mes"`
	Ano          int       `gorm:"not null" json:"ano"`
	Tipo         string    `gorm:"not null" json:"tipo"`      // EX: "TRIBUTO_INSS", "FOLHA_PAGAMENTO", "ISS"
	Descricao    string    `json:"descricao"`
	Valor        int64     `json:"valor"`                     // Armazenado em centavos para precisão
	ReferenciaID uint      `json:"referenciaId"`              // ID da Holerite ou Guia Fiscal
	Status       string    `gorm:"default:'PENDENTE'" json:"status"` // PENDENTE, PAGO
	DataVencimento *time.Time `json:"dataVencimento"`
	CreatedAt    time.Time `json:"createdAt"`
}
