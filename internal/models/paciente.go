package models

import (
	"time"
	"odonto-flow-go/internal/utils"

	"gorm.io/gorm"
)

type Paciente struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    Nome             string         `gorm:"not null" json:"nome"`
    CPF              string         `gorm:"index" json:"cpf"`
    CPFHash          string         `gorm:"uniqueIndex" json:"-"`
    CodigoUnico      string         `gorm:"uniqueIndex" json:"codigoUnico"`
    RG               string         `json:"rg"`
    DataNascimento   time.Time      `json:"dataNascimento"`
    Genero           string         `json:"genero"` // M, F, O
    TelefonePrincipal string        `json:"telefonePrincipal"`
    Email            string         `json:"email"`
    EnderecoCompleto string         `json:"enderecoCompleto"`
    ClinicaID        uint           `gorm:"not null" json:"clinicaId"`
    Clinica          Clinica        `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    CreatedAt        time.Time      `json:"createdAt"`
    UpdatedAt        time.Time      `json:"updatedAt"`
    DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Paciente) BeforeSave(tx *gorm.DB) (err error) {
	if p.CPF != "" {
		// Gerar o Hash Determinístico para busca (Blind Index) ANTES de encriptar
		hash, err := utils.HashDeterministic(p.CPF)
		if err != nil {
			return err
		}
		p.CPFHash = hash

		p.CPF, err = utils.EncryptAESGCM(p.CPF)
		if err != nil {
			return err
		}
	}
	if p.RG != "" {
		p.RG, err = utils.EncryptAESGCM(p.RG)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Paciente) AfterFind(tx *gorm.DB) (err error) {
	if p.CPF != "" {
		p.CPF, err = utils.DecryptAESGCM(p.CPF)
		if err != nil {
			return err
		}
	}
	if p.RG != "" {
		p.RG, err = utils.DecryptAESGCM(p.RG)
		if err != nil {
			return err
		}
	}
	return nil
}
