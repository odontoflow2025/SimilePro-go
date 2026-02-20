package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlertaHandler struct {
	DB *gorm.DB
}

func NewAlertaHandler(db *gorm.DB) *AlertaHandler {
	return &AlertaHandler{DB: db}
}

// Create godoc
// @Summary      Create a health alert
// @Description  Create a critical health alert (e.g., Allergy, Diabetes) for a patient
// @Tags         alertas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string  true  "Patient ID"
// @Param        input  body      object  true  "Alert Info"
// @Success      201    {object}  models.Alerta
// @Router       /pacientes/{id}/alertas [post]
func (h *AlertaHandler) Create(c *gin.Context) {
	pacienteIDStr := c.Param("id")
	pacienteID, err := strconv.ParseUint(pacienteIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	// Verify if patient belongs to the user's clinic
	var paciente models.Paciente
	if err := h.DB.First(&paciente, uint(pacienteID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente not found"})
		return
	}
	if paciente.ClinicaID != userClinicaID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para criar alertas para este paciente"})
		return
	}

	var input struct {
		Nivel     string `json:"nivel" binding:"required"` // GRAVISSIMO, GRAVE, ATENCAO
		Tipo      string `json:"tipo"`
		Descricao string `json:"descricao" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	var userID uint
	switch v := userIDVal.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	}

	alerta := models.Alerta{
		PacienteID:  uint(pacienteID),
		Nivel:       models.NivelAlerta(input.Nivel),
		Tipo:        input.Tipo,
		Descricao:   input.Descricao,
		CriadoPorID: userID,
	}

	if err := h.DB.Create(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar alerta"})
		return
	}

	c.JSON(http.StatusCreated, alerta)
}

// FindByPaciente godoc
// @Summary      List health alerts
// @Description  Retrieve all health alerts for a specific patient
// @Tags         alertas
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {array}   models.Alerta
// @Router       /pacientes/{id}/alertas [get]
func (h *AlertaHandler) FindByPaciente(c *gin.Context) {
	pacienteID := c.Param("id")
	userClinicaID, _ := c.Get("clinicaID")

	// Verify patient clinic before showing alertas
	var paciente models.Paciente
	if err := h.DB.First(&paciente, pacienteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente not found"})
		return
	}

	if userClinicaID == nil || paciente.ClinicaID != userClinicaID.(uint) {
		// Restricted view: only critical alerts if searched?
		// Actually, FindByPaciente is usually for own clinic.
		// For cross-clinic, FindOne already handles restricted alerts.
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	var alertas []models.Alerta
	if err := h.DB.Where("paciente_id = ?", pacienteID).Order("created_at DESC").Find(&alertas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar alertas"})
		return
	}
	c.JSON(http.StatusOK, alertas)
}

// Delete godoc
// @Summary      Remove health alert
// @Description  Delete a health alert by its ID
// @Tags         alertas
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      string  true  "Patient ID"
// @Param        alertaId  path      string  true  "Alert ID"
// @Success      200       {object}  map[string]string
// @Router       /pacientes/{id}/alertas/{alertaId} [delete]
func (h *AlertaHandler) Delete(c *gin.Context) {
	alertaID := c.Param("alertaId")
	userClinicaID, _ := c.Get("clinicaID")

	var alerta models.Alerta
	if err := h.DB.Preload("Paciente").First(&alerta, alertaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alerta not found"})
		return
	}

	if userClinicaID == nil || alerta.Paciente.ClinicaID != userClinicaID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	if err := h.DB.Delete(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar alerta"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alerta removido"})
}
