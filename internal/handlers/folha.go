package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FolhaHandler struct {
	DB      *gorm.DB
	Service *services.FolhaService
}

func NewFolhaHandler(db *gorm.DB) *FolhaHandler {
	return &FolhaHandler{
		DB:      db,
		Service: services.NewFolhaService(db),
	}
}

// Helper para obter clinicaID do contexto
func (h *FolhaHandler) getClinicaID(c *gin.Context) (uint, error) {
	return getClinicaIDFromContext(c)
}

// Request payload
type ProcessarFolhaReq struct {
	Mes int `json:"mes"`
	Ano int `json:"ano"`
}

// Rubricas
func (h *FolhaHandler) CreateRubrica(c *gin.Context) {
	clinicaID, _ := h.getClinicaID(c)
	var rubrica models.RubricaFolha
	if err := c.ShouldBindJSON(&rubrica); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	if err := h.Service.CreateRubrica(clinicaID, &rubrica); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar rubrica"})
		return
	}

	c.JSON(http.StatusCreated, rubrica)
}

func (h *FolhaHandler) GetRubricas(c *gin.Context) {
	clinicaID, _ := h.getClinicaID(c)
	rubricas, err := h.Service.GetRubricas(clinicaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar rubricas"})
		return
	}
	c.JSON(http.StatusOK, rubricas)
}

// ProcessarFolha refatorado para usar o serviço
func (h *FolhaHandler) ProcessarFolha(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	var req ProcessarFolhaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida"})
		return
	}

	competencia, err := h.Service.ProcessarFechamentoMes(clinicaID, req.Mes, req.Ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folha processada com sucesso", "competencia": competencia})
}

// GetHolerites retorna a lista de holerites para a competência dada
func (h *FolhaHandler) GetHolerites(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	mes, _ := strconv.Atoi(c.Query("mes"))
	ano, _ := strconv.Atoi(c.Query("ano"))

	var competencia models.CompetenciaFolha
	if err := h.DB.Where("clinica_id = ? AND mes = ? AND ano = ?", clinicaID, mes, ano).First(&competencia).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folha não encontrada para o período"})
		return
	}

	var holerites []models.Holerite
	h.DB.Preload("Funcionario").Preload("Eventos.Rubrica").Where("competencia_folha_id = ?", competencia.ID).Find(&holerites)

	c.JSON(http.StatusOK, gin.H{"competencia": competencia, "holerites": holerites})
}

