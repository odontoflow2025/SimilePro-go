package models

import (
	"time"

	"gorm.io/gorm"
)

type TipoMovimentacao string

const (
    MovimentacaoEntrada TipoMovimentacao = "ENTRADA"
    MovimentacaoSaida   TipoMovimentacao = "SAIDA"
    MovimentacaoAjuste  TipoMovimentacao = "AJUSTE"
)

type Produto struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    ClinicaID       uint           `gorm:"not null" json:"clinicaId"`
    Clinica         Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Nome            string         `gorm:"not null" json:"nome"`
    Descricao       string         `json:"descricao"`
    SKU             string         `json:"sku"`
    UnidadeMedida   string         `json:"unidadeMedida"` // Ex: "UN", "ML", "CX"
    PrecoCusto      float64        `json:"precoCusto"`
    EstoqueAtual    float64        `gorm:"default:0" json:"estoqueAtual"`
    EstoqueMinimo   float64        `gorm:"default:0" json:"estoqueMinimo"`
    CreatedAt       time.Time      `json:"createdAt"`
    UpdatedAt       time.Time      `json:"updatedAt"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type NotaFiscalEntrada struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    ClinicaID       uint           `gorm:"not null" json:"clinicaId"`
    Clinica         Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Numero          string         `gorm:"not null" json:"numero"`
    Serie           string         `json:"serie"`
    ChaveAcesso     string         `json:"chaveAcesso"`
    Fornecedor      string         `json:"fornecedor"`
    DataEmissao     time.Time      `json:"dataEmissao"`
    ValorTotal      float64        `json:"valorTotal"`
    TransacaoID     *uint          `json:"transacaoId"` // Relacionamento com o financeiro
    Transacao       *Transacao     `gorm:"foreignKey:TransacaoID" json:"transacao,omitempty"`
    Itens           []ItemNF       `gorm:"foreignKey:NotaFiscalID" json:"itens"`
    CreatedAt       time.Time      `json:"createdAt"`
    UpdatedAt       time.Time      `json:"updatedAt"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type ItemNF struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    NotaFiscalID    uint           `json:"notaFiscalId"`
    ProdutoID       uint           `json:"produtoId"`
    Produto         Produto        `gorm:"foreignKey:ProdutoID" json:"produto,omitempty"`
    Quantidade      float64        `json:"quantidade"`
    PrecoUnitario   float64        `json:"precoUnitario"`
    Subtotal        float64        `json:"subtotal"`
}

type MovimentacaoEstoque struct {
    ID              uint             `gorm:"primaryKey" json:"id"`
    ClinicaID       uint             `gorm:"not null" json:"clinicaId"`
    ProdutoID       uint             `json:"produtoId"`
    Produto         Produto          `gorm:"foreignKey:ProdutoID" json:"produto,omitempty"`
    Tipo            TipoMovimentacao `json:"tipo"`
    Quantidade      float64          `json:"quantidade"`
    NotaFiscalID    *uint            `json:"notaFiscalId"`
    Observacao      string           `json:"observacao"`
    CreatedAt       time.Time        `json:"createdAt"`
}
