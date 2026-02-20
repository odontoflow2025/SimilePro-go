package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EvolucaoHandler struct {
    DB *gorm.DB
}

func NewEvolucaoHandler(db *gorm.DB) *EvolucaoHandler {
    return &EvolucaoHandler{DB: db}
}

type CreateEvolucaoInput struct {
    PacienteID              uint      `json:"pacienteId" binding:"required"`
    DentistaID              uint      `json:"dentistaId" binding:"required"`
    ClinicaID               uint      `json:"clinicaId" binding:"required"`
    Descricao               string    `json:"descricao" binding:"required"`
    Data                    time.Time `json:"data" binding:"required"`
    ProcedimentoRealizadoID *uint     `json:"procedimentoRealizadoId"`
    ItemPlanoID             *uint     `json:"itemPlanoId"`
}

// Create godoc
// @Summary      Record clinical evolution
// @Description  Add a new clinical note for a patient's treatment session
// @Tags         evolucoes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateEvolucaoInput  true  "Evolution Info"
// @Success      201    {object}  models.Evolucao
// @Router       /evolucoes [post]
func (h *EvolucaoHandler) Create(c *gin.Context) {
	var input CreateEvolucaoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evolucao := models.Evolucao{
		PacienteID:              input.PacienteID,
		DentistaID:              input.DentistaID,
		ClinicaID:               input.ClinicaID,
		Descricao:               input.Descricao,
		Data:                    input.Data,
		ProcedimentoRealizadoID: input.ProcedimentoRealizadoID,
		ItemPlanoID:             input.ItemPlanoID,
	}

	if err := h.DB.Create(&evolucao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create evolucao"})
		return
	}

	// If linked to a plan item, we could mark it as done here (complex logic omitted for now)

	c.JSON(http.StatusCreated, evolucao)
}

// FindAll godoc
// @Summary      List evolutions
// @Description  Retrieve a chronological list of clinical notes for a patient or clinic
// @Tags         evolucoes
// @Produce      json
// @Security     BearerAuth
// @Param        pacienteId query     string  false  "Filter by Patient ID"
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Success      200        {array}   models.Evolucao
// @Router       /evolucoes [get]
func (h *EvolucaoHandler) FindAll(c *gin.Context) {
	pacienteID := c.Query("pacienteId")
	clinicaID := c.Query("clinicaId")

	var evolucoes []models.Evolucao
	query := h.DB.Preload("Paciente").Preload("Dentista").Preload("ProcedimentoRealizado")

	if pacienteID != "" {
		query = query.Where("paciente_id = ?", pacienteID)
	}
	if clinicaID != "" {
		query = query.Where("clinica_id = ?", clinicaID)
	}

	if err := query.Order("data desc").Find(&evolucoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch evolucoes"})
		return
	}

	c.JSON(http.StatusOK, evolucoes)
}
