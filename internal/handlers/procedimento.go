package handlers

import (
	"net/http"
	"SimilePro-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProcedimentoHandler struct {
    DB *gorm.DB
}

func NewProcedimentoHandler(db *gorm.DB) *ProcedimentoHandler {
    return &ProcedimentoHandler{DB: db}
}

type CreateProcedimentoInput struct {
    Nome            string  `json:"nome" binding:"required"`
    Codigo          string  `json:"codigo"`
    ValorReferencia float64 `json:"valorReferencia"`
    Ativo           bool    `json:"ativo"`
}

// Create godoc
// @Summary      Create a procedure
// @Description  Add a new dental procedure to the clinic's catalog
// @Tags         procedimentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateProcedimentoInput  true  "Procedure Info"
// @Success      201    {object}  models.Procedimento
// @Router       /procedimentos [post]
func (h *ProcedimentoHandler) Create(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input CreateProcedimentoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	procedimento := models.Procedimento{
		Nome:            input.Nome,
		Codigo:          input.Codigo,
		ValorReferencia: input.ValorReferencia,
		Ativo:           input.Ativo,
		ClinicaID:       userClinicaID,
	}

	if err := h.DB.Create(&procedimento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create procedimento"})
		return
	}

	c.JSON(http.StatusCreated, procedimento)
}

// FindAll godoc
// @Summary      List procedures
// @Description  Retrieve the catalog of dental procedures for the clinic
// @Tags         procedimentos
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Procedimento
// @Router       /procedimentos [get]
func (h *ProcedimentoHandler) FindAll(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	var procedimentos []models.Procedimento
	query := h.DB.Where("clinica_id = ?", userClinicaID)

	if err := query.Find(&procedimentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch procedimentos"})
		return
	}

	c.JSON(http.StatusOK, procedimentos)
}
