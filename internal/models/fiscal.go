package models

import (
	"time"

	"gorm.io/gorm"
)

type StatusNFS string

const (
	StatusNFSProcessando StatusNFS = "PROCESSANDO"
	StatusNFSEmitida     StatusNFS = "EMITIDA"
	StatusNFSErro        StatusNFS = "ERRO"
	StatusNFSCancelada   StatusNFS = "CANCELADA"
)

// Nota Fiscal de Serviço Eletrônica (Ligada a uma Fatura enviada ao paciente)
type NotaFiscalServico struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClinicaID uint      `gorm:"not null" json:"clinicaId"`
	Clinica   Clinica   `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	FaturaID  *uint     `json:"faturaId"`
	Fatura    *Fatura   `gorm:"foreignKey:FaturaID" json:"fatura,omitempty"`
	Status    StatusNFS `gorm:"default:PROCESSANDO" json:"status"`
	
	Numero        string     `json:"numero"`
	Rps           string     `json:"rps"` // Recibo Provisório de Serviços
	Serie         string     `json:"serie"`
	IssRetido     bool       `json:"issRetido"`
	ValorTotal    float64    `gorm:"not null" json:"valorTotal"`
	ValorDeduzido float64    `json:"valorDeduzido"`
	ValorBaseCalc float64    `json:"valorBaseCalc"`
	Aliquota      float64    `json:"aliquota"`
	ValorIss      float64    `json:"valorIss"`
	
	CodigoServico string     `json:"codigoServico"`
	ItemLc116     string     `json:"itemLc116"` // LC 116/2003 (ex: "4.01 - Medicina e biomedicina...")
	Descricao     string     `gorm:"type:text" json:"descricao"`
	
	UrlPdf        string     `json:"urlPdf"`
	UrlXml        string     `json:"urlXml"`
	Protocolo     string     `json:"protocolo"` // Chave da SEFAZ/Prefeitura
	AutorizacaoAt *time.Time `json:"autorizacaoAt"`
	
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Registro consolidado de Impostos (Guias de DAS, DARF PIS/COFINS) a serem pagas no mês
type ImpostoGuia struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ClinicaID      uint           `gorm:"not null" json:"clinicaId"`
	Clinica        Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Descricao      string         `gorm:"not null" json:"descricao"` // Ex: "DAS SIMPLES MÊS 05", "IRRF SÓCIOS"
	Tipo           string         `json:"tipo"` // DAS, DARF, GPS, ISS
	Competencia    string         `json:"competencia"` // Ex: "05/2026"
	ValorTotal     float64        `gorm:"not null" json:"valorTotal"`
	DataVencimento time.Time      `json:"dataVencimento"`
	PeriodoApurado string         `json:"periodoApurado"`
	Status         StatusFatura   `gorm:"default:PENDENTE" json:"status"` // Reutilizando status de fatura/transação
	TransacaoID    *uint          `json:"transacaoId"` // Relaciona com a Despesa Paga
	Transacao      *Transacao     `gorm:"foreignKey:TransacaoID" json:"transacao,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
