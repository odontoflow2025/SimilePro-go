package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgendamentoHandler struct {
    DB *gorm.DB
}

func NewAgendamentoHandler(db *gorm.DB) *AgendamentoHandler {
    return &AgendamentoHandler{DB: db}
}

type CreateAgendamentoInput struct {
    PacienteID       uint      `json:"pacienteId" binding:"required"`
    DentistaID       uint      `json:"dentistaId" binding:"required"`
    DataHoraInicio   time.Time `json:"dataHoraInicio" binding:"required"`
    DataHoraFim      time.Time `json:"dataHoraFim" binding:"required"`
    Motivo           string    `json:"motivo"`
    UsuarioCriacaoID uint      `json:"usuarioCriacaoId"`
}

type UpdateStatusInput struct {
    Status models.StatusAgendamento `json:"status" binding:"required"`
}

// Create godoc
// @Summary      Create an appointment
// @Description  Schedule a new appointment for a patient in the user's clinic
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateAgendamentoInput  true  "Appointment Info"
// @Success      201    {object}  models.Agendamento
// @Router       /agendamentos [post]
func (h *AgendamentoHandler) Create(c *gin.Context) {
	var input CreateAgendamentoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	// TODO: Validate time conflicts

	agendamento := models.Agendamento{
		PacienteID:       input.PacienteID,
		DentistaID:       input.DentistaID,
		ClinicaID:        userClinicaID.(uint),
		DataHoraInicio:   input.DataHoraInicio,
		DataHoraFim:      input.DataHoraFim,
		Motivo:           input.Motivo,
		UsuarioCriacaoID: input.UsuarioCriacaoID,
		Status:           models.StatusAgendamentoAgendado,
	}

	if err := h.DB.Create(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agendamento"})
		return
	}

	c.JSON(http.StatusCreated, agendamento)
}

// FindAll godoc
// @Summary      List appointments
// @Description  List appointments with date and dentist filtering (clinic-scoped)
// @Tags         agendamentos
// @Produce      json
// @Security     BearerAuth
// @Param        dataInicio  query     string  false  "Start Date (YYYY-MM-DD)"
// @Param        dataFim     query     string  false  "End Date (YYYY-MM-DD)"
// @Param        dentistaId  query     string  false  "Dentist ID"
// @Success      200         {array}   models.Agendamento
// @Router       /agendamentos [get]
func (h *AgendamentoHandler) FindAll(c *gin.Context) {
	userClinicaID, _ := c.Get("clinicaID")
	dataInicio := c.Query("dataInicio")
	dataFim := c.Query("dataFim")
	dentistaID := c.Query("dentistaId")

	var agendamentos []models.Agendamento
	query := h.DB.Preload("Paciente").Preload("Dentista").Preload("Dentista.Usuario")

	if userClinicaID != nil {
		query = query.Where("clinica_id = ?", userClinicaID)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	if dentistaID != "" {
		query = query.Where("dentista_id = ?", dentistaID)
	}
	if dataInicio != "" {
		query = query.Where("data_hora_inicio >= ?", dataInicio)
	}
	if dataFim != "" {
		query = query.Where("data_hora_fim <= ?", dataFim)
	}

	if err := query.Find(&agendamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agendamentos"})
		return
	}

	c.JSON(http.StatusOK, agendamentos)
}

// FindOne godoc
// @Summary      Get appointment details
// @Description  Get details of a specific appointment (clinic-scoped)
// @Tags         agendamentos
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Appointment ID"
// @Success      200  {object}  models.Agendamento
// @Failure      403  {object}  map[string]string
// @Router       /agendamentos/{id} [get]
func (h *AgendamentoHandler) FindOne(c *gin.Context) {
    id := c.Param("id")
    userClinicaID, _ := c.Get("clinicaID")

    var agendamento models.Agendamento
    if err := h.DB.Preload("Paciente").Preload("Dentista").First(&agendamento, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
        return
    }

    if userClinicaID != nil && agendamento.ClinicaID != userClinicaID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: este agendamento pertence a outra clínica"})
        return
    }

    c.JSON(http.StatusOK, agendamento)
}

// UpdateStatus godoc
// @Summary      Update appointment status
// @Description  Change the status of an appointment (e.g., Confirmed, Cancelled, Completed)
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Appointment ID"
// @Param        input  body      UpdateStatusInput  true  "New Status"
// @Success      200    {object}  models.Agendamento
// @Router       /agendamentos/{id}/status [patch]
func (h *AgendamentoHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	userClinicaID, _ := c.Get("clinicaID")

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agendamento models.Agendamento
	if err := h.DB.First(&agendamento, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if userClinicaID != nil && agendamento.ClinicaID != userClinicaID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: este agendamento pertence a outra clínica"})
		return
	}

	agendamento.Status = input.Status
	if err := h.DB.Save(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, agendamento)
}
