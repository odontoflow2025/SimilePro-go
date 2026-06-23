package handlers

import (
	"net/http"
	"strconv"

	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/repository"

	"github.com/gin-gonic/gin"
)

type ConfiguracaoHandler struct {
	Repo *repository.ConfiguracaoRepository
}

func NewConfiguracaoHandler(repo *repository.ConfiguracaoRepository) *ConfiguracaoHandler {
	return &ConfiguracaoHandler{Repo: repo}
}

// helper function to extract tenant ID safely
func getTenantID(c *gin.Context) (uint, bool) {
	clinicaIDRaw, exists := c.Get("clinicaID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID não encontrado no contexto. O usuário está autenticado?"})
		return 0, false
	}
	
	// Handles both float64 (from JWT standard unmarshaling) and uint
	var clinicaID uint
	switch v := clinicaIDRaw.(type) {
	case float64:
		clinicaID = uint(v)
	case uint:
		clinicaID = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo inválido para clinica_id"})
		return 0, false
	}
	return clinicaID, true
}

// GET /api/v1/configuracoes
// Carrega todas as configurações agregadas (Escalares, Salas, Categorias e Integrações)
func (h *ConfiguracaoHandler) GetAgregado(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	configGeral, err := h.Repo.GetConfiguracaoClinica(clinicaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar configurações escalares"})
		return
	}

	salas, _ := h.Repo.ListSalas(clinicaID)
	integracoes, _ := h.Repo.ListIntegracoes(clinicaID)
	categorias, _ := h.Repo.ListCategoriasFornecimento(clinicaID)
	canais, _ := h.Repo.ListCanaisComunicacao(clinicaID)

	c.JSON(http.StatusOK, gin.H{
		"geral":       configGeral,
		"salas":       salas,
		"integracoes": integracoes,
		"categorias":  categorias,
		"canais":      canais,
	})
}

// PUT /api/v1/configuracoes/geral
func (h *ConfiguracaoHandler) UpdateGeral(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	var payload models.ConfiguracaoClinica
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido", "details": err.Error()})
		return
	}

	// Força o isolamento por tenant no payload, independentemente do que o client mandou
	payload.ClinicaID = clinicaID

	if err := h.Repo.SaveConfiguracaoClinica(&payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar configuração geral"})
		return
	}

	c.JSON(http.StatusOK, payload)
}

// POST /api/v1/configuracoes/salas
func (h *ConfiguracaoHandler) CreateSala(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	var sala models.SalaAtendimento
	if err := c.ShouldBindJSON(&sala); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	sala.ID = 0 // Garante nova inserção
	sala.ClinicaID = clinicaID

	if err := h.Repo.CreateSala(&sala); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar sala"})
		return
	}

	c.JSON(http.StatusCreated, sala)
}

// PUT /api/v1/configuracoes/salas/:id
func (h *ConfiguracaoHandler) UpdateSala(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	salaID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da sala inválido"})
		return
	}

	var sala models.SalaAtendimento
	if err := c.ShouldBindJSON(&sala); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	sala.ID = uint(salaID)
	sala.ClinicaID = clinicaID

	// Usar o método recém criado no repositório para salvar a sala de forma isolada e segura
	if err := h.Repo.SaveSala(&sala); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar sala"})
		return
	}

	c.JSON(http.StatusOK, sala)
}

// DELETE /api/v1/configuracoes/salas/:id
func (h *ConfiguracaoHandler) DeleteSala(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	salaID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.Repo.DeleteSala(clinicaID, uint(salaID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover sala"})
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /api/v1/configuracoes/integracoes
func (h *ConfiguracaoHandler) CreateIntegracao(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	var integracao models.IntegracaoClinica
	if err := c.ShouldBindJSON(&integracao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	integracao.ID = 0
	integracao.ClinicaID = clinicaID

	if err := h.Repo.SaveIntegracao(&integracao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar integração"})
		return
	}

	c.JSON(http.StatusCreated, integracao)
}

// PUT /api/v1/configuracoes/integracoes/:id
func (h *ConfiguracaoHandler) UpdateIntegracao(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	integracaoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var integracao models.IntegracaoClinica
	if err := c.ShouldBindJSON(&integracao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	integracao.ID = uint(integracaoID)
	integracao.ClinicaID = clinicaID

	if err := h.Repo.SaveIntegracao(&integracao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar integração"})
		return
	}

	c.JSON(http.StatusOK, integracao)
}

// DELETE /api/v1/configuracoes/integracoes/:id
func (h *ConfiguracaoHandler) DeleteIntegracao(c *gin.Context) {
	clinicaID, ok := getTenantID(c)
	if !ok {
		return
	}

	integracaoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.Repo.DeleteIntegracao(clinicaID, uint(integracaoID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover integração"})
		return
	}

	c.Status(http.StatusNoContent)
}
