package models

import (
	"time"

	"gorm.io/gorm"
)

type Clinica struct {
    ID                   uint           `gorm:"primaryKey" json:"id"`
    NomeFantasia         string         `gorm:"not null" json:"nomeFantasia"`
    RazaoSocial          string         `gorm:"not null" json:"razaoSocial"`
    CNPJ                 string         `gorm:"uniqueIndex;not null" json:"cnpj"`
    Logo                 string         `json:"logo"`
    Endereco             string         `json:"endereco"`
    Telefone             string         `json:"telefone"`
    ResponsavelID        uint           `json:"responsavelId"`
    Responsavel          *User          `gorm:"foreignKey:ResponsavelID" json:"responsavel,omitempty"`
    MatrizID             *uint          `json:"matrizId"`
    Matriz               *Clinica       `gorm:"foreignKey:MatrizID" json:"matriz,omitempty"`
    RndsAtivo            bool           `json:"rndsAtivo" gorm:"default:false"`
    RndsCnes             string         `json:"rndsCnes"`
    Plano                string         `json:"plano" gorm:"default:'FREE'"` // FREE, BASE, PREMIUM, EDUCACIONAL
    PlanoExpiracao       *time.Time     `json:"planoExpiracao"`
    PlanoStatus          string         `json:"planoStatus" gorm:"default:'ACTIVE'"` // ACTIVE, EXPIRED, CANCELED
    PlanoUltimoPagamento *time.Time     `json:"planoUltimoPagamento"`
    MaxFuncionarios      int            `json:"maxFuncionarios" gorm:"default:2"`
    RndsCertificado      string         `json:"-"` // Don't expose certificate in JSON
    RndsSenhaCertificado string         `json:"-"` // Don't expose password
    CreatedAt            time.Time      `json:"createdAt"`
    UpdatedAt            time.Time      `json:"updatedAt"`
    DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}
