package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlanoTratamentoHandler struct {
    DB *gorm.DB
}

func NewPlanoTratamentoHandler(db *gorm.DB) *PlanoTratamentoHandler {
    return &PlanoTratamentoHandler{DB: db}
}

type CreateItemPlanoInput struct {
    ProcedimentoID uint    `json:"procedimentoId" binding:"required"`
    ValorUnitario  float64 `json:"valorUnitario" binding:"required"`
    Quantidade     int     `json:"quantidade" binding:"required"`
    DenteRegiao    string  `json:"denteRegiao"`
}

type CreatePlanoInput struct {
    PacienteID uint                   `json:"pacienteId" binding:"required"`
    DentistaID uint                   `json:"dentistaId" binding:"required"`
    ClinicaID  uint                   `json:"clinicaId" binding:"required"`
    Itens      []CreateItemPlanoInput `json:"itens" binding:"dive"`
}

type UpdateStatusPlanoInput struct {
    Status models.StatusPlano `json:"status" binding:"required"`
}

// Create godoc
// @Summary      Create a treatment plan
// @Description  Create a new treatment plan with multiple procedure items
// @Tags         planos-tratamento
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreatePlanoInput  true  "Treatment Plan Info"
// @Success      201    {object}  models.PlanoTratamento
// @Router       /planos-tratamento [post]
func (h *PlanoTratamentoHandler) Create(c *gin.Context) {
	var input CreatePlanoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var valorTotal float64
	var itens []models.ItemPlano

	for _, itemInput := range input.Itens {
		subtotal := itemInput.ValorUnitario * float64(itemInput.Quantidade)
		valorTotal += subtotal
		itens = append(itens, models.ItemPlano{
			ProcedimentoID: itemInput.ProcedimentoID,
			ValorUnitario:  itemInput.ValorUnitario,
			Quantidade:     itemInput.Quantidade,
			DenteRegiao:    itemInput.DenteRegiao,
		})
	}

	plano := models.PlanoTratamento{
		PacienteID: input.PacienteID,
		DentistaID: input.DentistaID,
		ClinicaID:  input.ClinicaID,
		Status:     models.StatusPlanoOrcamento,
		ValorTotal: valorTotal,
		Itens:      itens,
	}

	if err := h.DB.Create(&plano).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plano de tratamento"})
		return
	}

	c.JSON(http.StatusCreated, plano)
}

// FindAll godoc
// @Summary      List treatment plans
// @Description  Retrieve a list of treatment plans with clinic and patient filtering
// @Tags         planos-tratamento
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Param        pacienteId query     string  false  "Filter by Patient ID"
// @Success      200        {array}   models.PlanoTratamento
// @Router       /planos-tratamento [get]
func (h *PlanoTratamentoHandler) FindAll(c *gin.Context) {
	clinicaID := c.Query("clinicaId")
	pacienteID := c.Query("pacienteId")

	var planos []models.PlanoTratamento
	query := h.DB.Preload("Paciente").Preload("Dentista").Preload("Itens").Preload("Itens.Procedimento")

	if clinicaID != "" {
		query = query.Where("clinica_id = ?", clinicaID)
	}
	if pacienteID != "" {
		query = query.Where("paciente_id = ?", pacienteID)
	}

	if err := query.Find(&planos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch planos"})
		return
	}

	c.JSON(http.StatusOK, planos)
}

// FindOne godoc
// @Summary      Get treatment plan details
// @Description  Retrieve detailed information about a specific treatment plan
// @Tags         planos-tratamento
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Plan ID"
// @Success      200  {object}  models.PlanoTratamento
// @Failure      404  {object}  map[string]string
// @Router       /planos-tratamento/{id} [get]
func (h *PlanoTratamentoHandler) FindOne(c *gin.Context) {
	id := c.Param("id")
	var plano models.PlanoTratamento
	if err := h.DB.Preload("Paciente").Preload("Dentista").Preload("Itens").Preload("Itens.Procedimento").First(&plano, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plano not found"})
		return
	}

	c.JSON(http.StatusOK, plano)
}

// UpdateStatus godoc
// @Summary      Update plan status
// @Description  Change the status of a treatment plan (e.g., from Draft to Approved)
// @Tags         planos-tratamento
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string                  true  "Plan ID"
// @Param        input  body      UpdateStatusPlanoInput  true  "New Status"
// @Success      200    {object}  models.PlanoTratamento
// @Router       /planos-tratamento/{id}/status [patch]
func (h *PlanoTratamentoHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var input UpdateStatusPlanoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plano models.PlanoTratamento
	if err := h.DB.First(&plano, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plano not found"})
		return
	}

	plano.Status = input.Status
	if err := h.DB.Save(&plano).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, plano)
}
