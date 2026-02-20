package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FaturaHandler struct {
    DB *gorm.DB
}

func NewFaturaHandler(db *gorm.DB) *FaturaHandler {
    return &FaturaHandler{DB: db}
}

type CreateFaturaInput struct {
    ClinicaID         uint      `json:"clinicaId" binding:"required"`
    PacienteID        uint      `json:"pacienteId" binding:"required"`
    PlanoTratamentoID *uint     `json:"planoTratamentoId"`
    ValorTotal        float64   `json:"valorTotal" binding:"required"`
    DataVencimento    time.Time `json:"dataVencimento" binding:"required"`
}

// Create godoc
// @Summary      Create an invoice
// @Description  Generate a new invoice for a patient, optionally linked to a treatment plan
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateFaturaInput  true  "Invoice Info"
// @Success      201    {object}  models.Fatura
// @Router       /financeiro/faturas [post]
func (h *FaturaHandler) Create(c *gin.Context) {
	var input CreateFaturaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fatura := models.Fatura{
		ClinicaID:         input.ClinicaID,
		PacienteID:        input.PacienteID,
		PlanoTratamentoID: input.PlanoTratamentoID,
		ValorTotal:        input.ValorTotal,
		Status:            models.StatusFaturaPendente,
		DataGeracao:       time.Now(),
		DataVencimento:    input.DataVencimento,
	}

	if err := h.DB.Create(&fatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fatura"})
		return
	}

	c.JSON(http.StatusCreated, fatura)
}

// FindAll godoc
// @Summary      List invoices
// @Description  Retrieve a list of invoices with filtering by clinic, patient and status
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Param        pacienteId query     string  false  "Filter by Patient ID"
// @Param        status     query     string  false  "Filter by Status"
// @Success      200        {array}   models.Fatura
// @Router       /financeiro/faturas [get]
func (h *FaturaHandler) FindAll(c *gin.Context) {
	clinicaID := c.Query("clinicaId")
	pacienteID := c.Query("pacienteId")
	status := c.Query("status")

	var faturas []models.Fatura
	query := h.DB.Preload("Paciente").Preload("PlanoTratamento")

	if clinicaID != "" {
		query = query.Where("clinica_id = ?", clinicaID)
	}
	if pacienteID != "" {
		query = query.Where("paciente_id = ?", pacienteID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&faturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch faturas"})
		return
	}

	c.JSON(http.StatusOK, faturas)
}


// GetAtrasadas godoc
// @Summary      List overdue invoices
// @Description  Retrieve all invoices that are past their due date and remain unpaid
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Fatura
// @Router       /financeiro/faturas/atrasadas [get]
func (h *FaturaHandler) GetAtrasadas(c *gin.Context) {
	var faturas []models.Fatura
	now := time.Now()

	if err := h.DB.Preload("Paciente").Preload("PlanoTratamento").
		Where("status = ? AND data_vencimento < ?", models.StatusFaturaPendente, now).
		Find(&faturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overdue invoices"})
		return
	}

	c.JSON(http.StatusOK, faturas)
}
