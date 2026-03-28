package models

import (
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
    Salario      float64           `json:"salario"`
    PIS          string            `json:"pis"`
    CTPS         string            `json:"ctps"`
    RG           string            `json:"rg"`
    CargaHoraria int               `json:"cargaHoraria"` // Horas mensais
    Dependentes  int               `json:"dependentes"` // Qtd p/ cálculo IRRF/Salário Família
    DadosBanco   string            `json:"dadosBanco"`
    
    // Status e Demissão
    Status       StatusFuncionario `gorm:"default:ATIVO" json:"status"`
    DataDemissao *time.Time        `json:"dataDemissao"`
    MotivoDemissao string          `json:"motivoDemissao"`
    
    CreatedAt    time.Time         `json:"createdAt"`
    UpdatedAt    time.Time         `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt    `gorm:"index" json:"-"`
}
