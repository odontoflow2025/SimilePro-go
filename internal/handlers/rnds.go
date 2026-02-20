package handlers

import (
	"net/http"
	"odonto-flow-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RndsHandler struct {
	DB      *gorm.DB
	Service *services.RndsService
}

func NewRndsHandler(db *gorm.DB) *RndsHandler {
	return &RndsHandler{
		DB:      db,
		Service: services.NewRndsService(db),
	}
}

// CheckStatus verifies RNDS connection
// CheckStatus godoc
// @Summary      Verify RNDS connection
// @Description  Check if the clinic has a valid connection to the National Health Data Network
// @Tags         rnds
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /rnds/status [get]
func (h *RndsHandler) CheckStatus(c *gin.Context) {
	// Assume authenticated user belongs to a clinic
	// Getting clinicaID from user context or DB
	// For now, mock or assume 1 if not found, similar to other handlers
	clinicaID := uint(1) // TODO: Extract from context properly

	valid, msg, err := h.Service.CheckStatus(clinicaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected": valid,
		"message":   msg,
		"timestamp": "2024-01-01T00:00:00Z", // Mock
	})
}

// SendPatient sends a patient to RNDS
// SendPatient godoc
// @Summary      Sync patient with RNDS
// @Description  Send patient record to the National Health Data Network for national identification
// @Tags         rnds
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  map[string]string
// @Router       /rnds/paciente/{id} [post]
func (h *RndsHandler) SendPatient(c *gin.Context) {
	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	rndsID, err := h.Service.SendPatient(uint(patientID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Paciente enviado com sucesso",
		"rndsId":  rndsID,
	})
}

// SendAttendance sends an encounter
// SendAttendance godoc
// @Summary      Submit high summary (Sumário de Alta)
// @Description  Send a clinical attendance record/summary to RNDS
// @Tags         rnds
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Appointment ID"
// @Success      200  {object}  map[string]string
// @Router       /rnds/atendimento/{id} [post]
func (h *RndsHandler) SendAttendance(c *gin.Context) {
	agendamentoIDStr := c.Param("id")
	agendamentoID, err := strconv.ParseUint(agendamentoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	rndsID, err := h.Service.SendAttendance(uint(agendamentoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Atendimento enviado com sucesso",
		"rndsId":  rndsID,
	})
}
