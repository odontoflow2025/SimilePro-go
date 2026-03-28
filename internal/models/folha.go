package models

import (
	"time"

	"gorm.io/gorm"
)

type StatusFolha string
type TipoEventoFolha string

const (
	StatusFolhaAberta   StatusFolha = "ABERTA"
	StatusFolhaFechada  StatusFolha = "FECHADA"
	StatusFolhaPaga     StatusFolha = "PAGA"

	TipoEventoProvento TipoEventoFolha = "PROVENTO"
	TipoEventoDesconto TipoEventoFolha = "DESCONTO"
)

// Folha de Pagamento Consolidada (Geral por Clínica e Mês/Ano)
type CompetenciaFolha struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	ClinicaID uint        `gorm:"not null" json:"clinicaId"`
	Clinica   Clinica     `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Mes       int         `gorm:"not null" json:"mes"`
	Ano       int         `gorm:"not null" json:"ano"`
	Status    StatusFolha `gorm:"default:ABERTA" json:"status"`
	TotalBase float64     `json:"totalBase"`
	TotalLiq  float64     `json:"totalLiq"`
	TotalInss float64     `json:"totalInss"`
	TotalIrrf float64     `json:"totalIrrf"`
	TotalFgts float64     `json:"totalFgts"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Holerite individual do funcionário na competência
type Holerite struct {
	ID                 uint             `gorm:"primaryKey" json:"id"`
	CompetenciaFolhaID uint             `gorm:"not null" json:"competenciaFolhaId"`
	Competencia        CompetenciaFolha `gorm:"foreignKey:CompetenciaFolhaID" json:"competencia,omitempty"`
	FuncionarioID      uint             `gorm:"not null" json:"funcionarioId"`
	Funcionario        Funcionario      `gorm:"foreignKey:FuncionarioID" json:"funcionario,omitempty"`
	DiasTrabalhados    int              `json:"diasTrabalhados"`
	SalarioBase        float64          `json:"salarioBase"`
	TotalProventos     float64          `json:"totalProventos"`
	TotalDescontos     float64          `json:"totalDescontos"`
	SalarioLiquido     float64          `json:"salarioLiquido"`
	Eventos            []EventoHolerite `gorm:"foreignKey:HoleriteID" json:"eventos"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}

// Eventos específicos (Rubricas: DSR, INSS, Faltas, Horas Extras) dentro do Holerite
type EventoHolerite struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	HoleriteID uint            `gorm:"not null" json:"holeriteId"`
	Descricao  string          `gorm:"not null" json:"descricao"` // Ex: "INSS", "IRRF", "Vale Transporte"
	Tipo       TipoEventoFolha `gorm:"not null" json:"tipo"`
	Referencia string          `json:"referencia"`                // Ex: "11%", "5 dias", "2 horas"
	Valor      float64         `gorm:"not null" json:"valor"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}
