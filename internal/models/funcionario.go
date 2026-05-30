package models

import (
	"fmt"
	"odonto-flow-go/internal/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type StatusFuncionario string

const (
    StatusFuncionarioAtivo    StatusFuncionario = "ATIVO"
    StatusFuncionarioAfastado StatusFuncionario = "AFASTADO"
    StatusFuncionarioDemitido StatusFuncionario = "DEMITIDO"
)


type Funcionario struct {
    ID           uint              `gorm:"primaryKey" json:"id"`
    UsuarioID    uint              `gorm:"uniqueIndex:idx_user_clinica_fun;not null" json:"usuarioId"`
    Usuario      User              `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
    ClinicaID    uint              `gorm:"uniqueIndex:idx_user_clinica_fun;not null" json:"clinicaId"`
    Clinica      Clinica           `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Cargo        string            `json:"cargo"` // RECEPCIONISTA, ASB, TSB, FINANCEIRO, etc.
    
    // Admission / Trabalhista
    DataAdmissao time.Time         `json:"dataAdmissao"`
    Salario      float64           `gorm:"-" json:"salario"` // Campo em memória para cálculos
    SalarioDB    string            `gorm:"column:salario" json:"-"` // Campo criptografado no banco
    PIS          string            `json:"pis"`
    CTPS         string            `json:"ctps"`
    RG           string            `json:"rg"`
    CargaHoraria int               `json:"cargaHoraria"` // Horas mensais
    Dependentes  int               `json:"dependentes"` // Qtd p/ cálculo IRRF/Salário Família
    DadosBanco   string            `json:"dadosBanco"`
    CNPJ         string            `json:"cnpj"` // Opcional p/ contratação PJ
    Endereco     string            `json:"endereco"` // Logradouro, número, bairro, cidade, uf, cep
    
    // Status e Demissão
    Status       StatusFuncionario `gorm:"default:ATIVO" json:"status"`
    DataDemissao *time.Time        `json:"dataDemissao"`
    MotivoDemissao string          `json:"motivoDemissao"`
    
    CreatedAt    time.Time         `json:"createdAt"`
    UpdatedAt    time.Time         `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (f *Funcionario) BeforeSave(tx *gorm.DB) (err error) {
	// Encriptar Salário
	if f.Salario > 0 {
		f.SalarioDB, err = utils.EncryptAESGCM(fmt.Sprintf("%.2f", f.Salario))
		if err != nil {
			return err
		}
	}

	// Encriptar outros campos sensíveis
	if f.PIS != "" {
		f.PIS, _ = utils.EncryptAESGCM(f.PIS)
	}
	if f.CTPS != "" {
		f.CTPS, _ = utils.EncryptAESGCM(f.CTPS)
	}
	if f.RG != "" {
		f.RG, _ = utils.EncryptAESGCM(f.RG)
	}
	if f.DadosBanco != "" {
		f.DadosBanco, _ = utils.EncryptAESGCM(f.DadosBanco)
	}
	if f.CNPJ != "" {
		f.CNPJ, _ = utils.EncryptAESGCM(f.CNPJ)
	}
	if f.Endereco != "" {
		f.Endereco, _ = utils.EncryptAESGCM(f.Endereco)
	}

	return nil
}

func (f *Funcionario) AfterFind(tx *gorm.DB) (err error) {
	// Decriptar Salário
	if f.SalarioDB != "" {
		val, err := utils.DecryptAESGCM(f.SalarioDB)
		if err == nil {
			f.Salario, _ = strconv.ParseFloat(val, 64)
		}
	}

	// Decriptar outros campos
	f.PIS, _ = utils.DecryptAESGCM(f.PIS)
	f.CTPS, _ = utils.DecryptAESGCM(f.CTPS)
	f.RG, _ = utils.DecryptAESGCM(f.RG)
	f.DadosBanco, _ = utils.DecryptAESGCM(f.DadosBanco)
	f.CNPJ, _ = utils.DecryptAESGCM(f.CNPJ)
	f.Endereco, _ = utils.DecryptAESGCM(f.Endereco)

	return nil
}

