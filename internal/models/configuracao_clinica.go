package models

import (
	"encoding/json"
	"time"

	"SimilePro-go/internal/utils"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ConfiguracaoClinica agrupa configuracoes escalares (1:1) do tenant.
type ConfiguracaoClinica struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	ClinicaID uint    `gorm:"uniqueIndex;not null" json:"clinicaId"`
	Clinica   Clinica `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`

	// Preferências Gerais
	TemaSistema      string `gorm:"default:'light'" json:"temaSistema"`
	NotificarEmail   bool   `gorm:"default:true" json:"notificarEmail"`
	BackupAutomatico bool   `gorm:"default:true" json:"backupAutomatico"`

	// Portal do Paciente
	PortalAtivo            bool   `gorm:"default:false" json:"portalAtivo"`
	PermitePagamentoOnline bool   `gorm:"default:false" json:"permitePagamentoOnline"`
	RegrasReagendamento    string `gorm:"type:text" json:"regrasReagendamento"`
	MensagemBoasVindas     string `gorm:"type:text" json:"mensagemBoasVindas"`
	VisHistorico           bool   `gorm:"default:true" json:"visHistorico"`
	VisPrescricao          bool   `gorm:"default:true" json:"visPrescricao"`
	VisExames              bool   `gorm:"default:false" json:"visExames"`
	VisOdontograma         bool   `gorm:"default:false" json:"visOdontograma"`
	NotificaPacienteEmail  bool   `gorm:"default:true" json:"notificaPacienteEmail"`
	NotificaPacienteSMS    bool   `gorm:"default:false" json:"notificaPacienteSMS"`

	// Relatórios e BI
	AcessoDadosFinanceiros bool   `gorm:"default:false" json:"acessoDadosFinanceiros"`
	FiltroPeriodoPadrao    string `gorm:"default:'30d'" json:"filtroPeriodoPadrao"`

	// Configuração de Fornecedores Padrão
	PrazoPagtoFornecedorDias int  `gorm:"default:30" json:"prazoPagtoFornecedorDias"`
	ExigirNFFornecedor       bool `gorm:"default:true" json:"exigirNFFornecedor"`

	UpdatedAt time.Time `json:"updatedAt"`
}

// SalaAtendimento representa um consultório ou sala (1:N com Clinica)
type SalaAtendimento struct {
	ID                  uint    `gorm:"primaryKey" json:"id"`
	ClinicaID           uint    `gorm:"not null;index" json:"clinicaId"`
	Clinica             Clinica `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Nome                string  `gorm:"not null" json:"nome"`
	Equipamentos        string  `gorm:"type:text" json:"equipamentos"`
	ExigeLimpeza        bool    `gorm:"default:false" json:"exigeLimpeza"`
	TempoLimpezaMinutos int     `gorm:"default:0" json:"tempoLimpezaMinutos"`
}

// IntegracaoClinica representa integrações de terceiros (1:N com Clinica)
type IntegracaoClinica struct {
	ID                     uint              `gorm:"primaryKey" json:"id"`
	ClinicaID              uint              `gorm:"not null;index" json:"clinicaId"`
	Clinica                Clinica           `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Plataforma             string            `gorm:"not null" json:"plataforma"`
	Ativo                  bool              `gorm:"default:false" json:"ativo"`
	Credenciais            datatypes.JSONMap `gorm:"-" json:"credenciais"`                        // Exposto na API (in-memory), ignorado no BD
	CredenciaisEncriptadas string            `gorm:"column:credenciais;type:text" json:"-"`       // Criptografado com AES-GCM no BD
}

// GORM Hook: BeforeSave realiza a criptografia AES-GCM das Credenciais.
func (i *IntegracaoClinica) BeforeSave(tx *gorm.DB) (err error) {
	if len(i.Credenciais) > 0 {
		bytes, err := json.Marshal(i.Credenciais)
		if err != nil {
			return err
		}
		encrypted, err := utils.EncryptAESGCM(string(bytes))
		if err != nil {
			return err
		}
		i.CredenciaisEncriptadas = encrypted
	} else {
		i.CredenciaisEncriptadas = ""
	}
	return nil
}

// GORM Hook: AfterFind realiza a descriptografia AES-GCM das Credenciais para a camada de serviço.
func (i *IntegracaoClinica) AfterFind(tx *gorm.DB) (err error) {
	if i.CredenciaisEncriptadas != "" {
		decrypted, err := utils.DecryptAESGCM(i.CredenciaisEncriptadas)
		if err == nil && decrypted != "" {
			var creds datatypes.JSONMap
			if err := json.Unmarshal([]byte(decrypted), &creds); err == nil {
				i.Credenciais = creds
			}
		}
	}
	return nil
}

// CategoriaFornecimento (1:N com Clinica)
type CategoriaFornecimento struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	ClinicaID uint    `gorm:"not null;index" json:"clinicaId"`
	Clinica   Clinica `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Nome      string  `gorm:"not null" json:"nome"`
}

// CanalComunicacao (1:N com Clinica)
type CanalComunicacao struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	ClinicaID uint    `gorm:"not null;index" json:"clinicaId"`
	Clinica   Clinica `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Nome      string  `gorm:"not null" json:"nome"`
	Ativo     bool    `gorm:"default:true" json:"ativo"`
}

// ModeloAvisoInterno (1:N com Clinica)
type ModeloAvisoInterno struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	ClinicaID uint    `gorm:"not null;index" json:"clinicaId"`
	Clinica   Clinica `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	Mensagem  string  `gorm:"type:text;not null" json:"mensagem"`
}
