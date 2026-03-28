package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConvenioHandler struct {
    DB *gorm.DB
}

func NewConvenioHandler(db *gorm.DB) *ConvenioHandler {
    return &ConvenioHandler{DB: db}
}

type CreateConvenioInput struct {
    Nome         string `json:"nome" binding:"required"`
    RegistroANS  string `json:"registroAns"`
    TabelaPrecos string `json:"tabelaPrecos"`
    ClinicaID    uint   `json:"clinicaId" binding:"required"`
}

// Create godoc
// @Summary      Register health insurance
// @Description  Create a new health insurance company (convenio) association
// @Tags         convenios
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateConvenioInput  true  "Insurance Info"
// @Success      201    {object}  models.Convenio
// @Router       /convenios [post]
func (h *ConvenioHandler) Create(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input CreateConvenioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convenio := models.Convenio{
		Nome:         input.Nome,
		RegistroANS:  input.RegistroANS,
		TabelaPrecos: input.TabelaPrecos,
		ClinicaID:    userClinicaID, // Force current clinic ID
	}

	if err := h.DB.Create(&convenio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create convenio"})
		return
	}

	c.JSON(http.StatusCreated, convenio)
}

// FindAll godoc
// @Summary      List health insurances
// @Description  Retrieve all health insurance companies for a clinic
// @Tags         convenios
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Success      200        {array}   models.Convenio
// @Router       /convenios [get]
func (h *ConvenioHandler) FindAll(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	reqClinicaID := c.Query("clinicaId")

	var convenios []models.Convenio
	query := h.DB

	if userRole == "ADMIN_TOTAL" && reqClinicaID != "" {
		query = query.Where("clinica_id = ?", reqClinicaID)
	} else {
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	if err := query.Find(&convenios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch convenios"})
		return
	}

	c.JSON(http.StatusOK, convenios)
}
