package repository

import (
	"odonto-flow-go/internal/models"

	"gorm.io/gorm"
)

type ConfiguracaoRepository struct {
	db *gorm.DB
}

func NewConfiguracaoRepository(db *gorm.DB) *ConfiguracaoRepository {
	return &ConfiguracaoRepository{db: db}
}

// TenantScope é um GORM Scope genérico que impõe o isolamento de dados por clinica_id.
// Nenhuma operação no repositório deve ser feita sem aplicar este scope.
func TenantScope(clinicaID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinica_id = ?", clinicaID)
	}
}

// GetConfiguracaoClinica retorna as configurações escalares da clínica, ou cria uma nova com defaults se não existir.
func (r *ConfiguracaoRepository) GetConfiguracaoClinica(clinicaID uint) (*models.ConfiguracaoClinica, error) {
	var config models.ConfiguracaoClinica
	err := r.db.Scopes(TenantScope(clinicaID)).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Cria configuração padrão caso seja a primeira vez
			config = models.ConfiguracaoClinica{ClinicaID: clinicaID}
			if createErr := r.db.Create(&config).Error; createErr != nil {
				return nil, createErr
			}
			return &config, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *ConfiguracaoRepository) SaveConfiguracaoClinica(config *models.ConfiguracaoClinica) error {
	// Garante que o clinica_id seja preservado e respeitado
	return r.db.Scopes(TenantScope(config.ClinicaID)).Save(config).Error
}

func (r *ConfiguracaoRepository) ListSalas(clinicaID uint) ([]models.SalaAtendimento, error) {
	var salas []models.SalaAtendimento
	err := r.db.Scopes(TenantScope(clinicaID)).Find(&salas).Error
	return salas, err
}

func (r *ConfiguracaoRepository) CreateSala(sala *models.SalaAtendimento) error {
	return r.db.Create(sala).Error
}

func (r *ConfiguracaoRepository) SaveSala(sala *models.SalaAtendimento) error {
	return r.db.Scopes(TenantScope(sala.ClinicaID)).Save(sala).Error
}

func (r *ConfiguracaoRepository) DeleteSala(clinicaID uint, salaID uint) error {
	return r.db.Scopes(TenantScope(clinicaID)).Delete(&models.SalaAtendimento{}, salaID).Error
}

func (r *ConfiguracaoRepository) ListIntegracoes(clinicaID uint) ([]models.IntegracaoClinica, error) {
	var integracoes []models.IntegracaoClinica
	err := r.db.Scopes(TenantScope(clinicaID)).Find(&integracoes).Error
	return integracoes, err
}

func (r *ConfiguracaoRepository) SaveIntegracao(integracao *models.IntegracaoClinica) error {
	return r.db.Save(integracao).Error
}

func (r *ConfiguracaoRepository) DeleteIntegracao(clinicaID uint, integracaoID uint) error {
	return r.db.Scopes(TenantScope(clinicaID)).Delete(&models.IntegracaoClinica{}, integracaoID).Error
}

func (r *ConfiguracaoRepository) ListCategoriasFornecimento(clinicaID uint) ([]models.CategoriaFornecimento, error) {
	var categorias []models.CategoriaFornecimento
	err := r.db.Scopes(TenantScope(clinicaID)).Find(&categorias).Error
	return categorias, err
}

func (r *ConfiguracaoRepository) ListCanaisComunicacao(clinicaID uint) ([]models.CanalComunicacao, error) {
	var canais []models.CanalComunicacao
	err := r.db.Scopes(TenantScope(clinicaID)).Find(&canais).Error
	return canais, err
}
