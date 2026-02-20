package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnamneseHandler struct {
    DB *gorm.DB
}

func NewAnamneseHandler(db *gorm.DB) *AnamneseHandler {
    return &AnamneseHandler{DB: db}
}

type CreateAnamneseInput struct {
    PacienteID      uint   `json:"pacienteId" binding:"required"`
    HistoricoMedico string `json:"historicoMedico"`
    Alergias        string `json:"alergias"`
    Medicamentos    string `json:"medicamentos"`
    Fumante         bool   `json:"fumante"`
    Gestante        bool   `json:"gestante"`
    Observacoes     string `json:"observacoes"`
}

// CreateOrUpdate godoc
// @Summary      Save patient anamnesis
// @Description  Create or update the health history for a patient
// @Tags         anamnese
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateAnamneseInput  true  "Anamnesis Info"
// @Success      200    {object}  models.Anamnese
// @Router       /pacientes/{id}/anamnese [post]
func (h *AnamneseHandler) CreateOrUpdate(c *gin.Context) {
	var input CreateAnamneseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var anamnese models.Anamnese
	// Check if exists for patient
	result := h.DB.Where("paciente_id = ?", input.PacienteID).First(&anamnese)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new
		anamnese = models.Anamnese{
			PacienteID:      input.PacienteID,
			HistoricoMedico: input.HistoricoMedico,
			Alergias:        input.Alergias,
			Medicamentos:    input.Medicamentos,
			Fumante:         input.Fumante,
			Gestante:        input.Gestante,
			Observacoes:     input.Observacoes,
		}
		if err := h.DB.Create(&anamnese).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create anamnese"})
			return
		}
	} else {
		// Update existing
		anamnese.HistoricoMedico = input.HistoricoMedico
		anamnese.Alergias = input.Alergias
		anamnese.Medicamentos = input.Medicamentos
		anamnese.Fumante = input.Fumante
		anamnese.Gestante = input.Gestante
		anamnese.Observacoes = input.Observacoes

		if err := h.DB.Save(&anamnese).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update anamnese"})
			return
		}
	}

	c.JSON(http.StatusOK, anamnese)
}

// FindByPaciente godoc
// @Summary      Get patient anamnesis
// @Description  Retrieve the medical record for a specific patient
// @Tags         anamnese
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  models.Anamnese
// @Failure      404  {object}  map[string]string
// @Router       /pacientes/{id}/anamnese [get]
func (h *AnamneseHandler) FindByPaciente(c *gin.Context) {
	pacienteID := c.Param("id")
	var anamnese models.Anamnese

	if err := h.DB.Where("paciente_id = ?", pacienteID).First(&anamnese).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Anamnese not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anamnese"})
		}
		return
	}

	c.JSON(http.StatusOK, anamnese)
}
