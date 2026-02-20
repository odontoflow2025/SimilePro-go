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
    CreatedAt      time.Time      `json:"createdAt"`
    UpdatedAt      time.Time      `json:"updatedAt"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
