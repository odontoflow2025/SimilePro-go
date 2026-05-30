package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DentistaHandler struct {
    DB *gorm.DB
}

func NewDentistaHandler(db *gorm.DB) *DentistaHandler {
    return &DentistaHandler{DB: db}
}

type CreateDentistaInput struct {
    UsuarioID          uint      `json:"usuarioId" binding:"required"`
    CRO                string    `json:"cro" binding:"required"`
    Especialidade      string    `json:"especialidade"`
    PorcentagemRepasse float64   `json:"porcentagemRepasse"`
    DataContratacao    time.Time `json:"dataContratacao"`
}

// Create godoc
// @Summary      Register a dentist
// @Description  Associate a new dentist with the user's clinic
// @Tags         dentistas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateDentistaInput  true  "Dentist Info"
// @Success      201    {object}  models.Dentista
// @Router       /dentistas [post]
func (h *DentistaHandler) Create(c *gin.Context) {
	var input CreateDentistaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	dentista := models.Dentista{
		UsuarioID:          input.UsuarioID,
		ClinicaID:          userClinicaID.(uint),
		CRO:                input.CRO,
		Especialidade:      input.Especialidade,
		PorcentagemRepasse: input.PorcentagemRepasse,
		DataContratacao:    input.DataContratacao,
	}

	if err := h.DB.Create(&dentista).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dentista"})
		return
	}

	c.JSON(http.StatusCreated, dentista)
}

// FindAll godoc
// @Summary      List dentists
// @Description  Retrieve a list of all dentists in the user's clinic
// @Tags         dentistas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Dentista
// @Router       /dentistas [get]
func (h *DentistaHandler) FindAll(c *gin.Context) {
	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	userRole, _ := c.Get("userRole")

	query := h.DB.Preload("Usuario")

	if userRole == "ADMIN_TOTAL" {
		reqClinicaID := c.Query("clinicaId")
		if reqClinicaID != "" {
			query = query.Where("clinica_id = ?", reqClinicaID)
		}
	} else {
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	var dentistas []models.Dentista
	if err := query.Find(&dentistas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dentistas"})
		return
	}

	c.JSON(http.StatusOK, dentistas)
}

// GetProfissionaisAgenda godoc
// @Summary      List professionals for agenda
// @Description  Retrieve a simplified list of professionals for the clinic's agenda
// @Tags         dentistas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  map[string]interface{}
// @Router       /v1/profissionais [get]
func (h *DentistaHandler) GetProfissionaisAgenda(c *gin.Context) {
	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	userRole, _ := c.Get("userRole")

	query := h.DB.Preload("Usuario")

	if userRole == "ADMIN_TOTAL" {
		reqClinicaID := c.Query("clinicaId")
		if reqClinicaID != "" {
			query = query.Where("clinica_id = ?", reqClinicaID)
		}
	} else {
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	var dentistas []models.Dentista
	if err := query.Find(&dentistas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profissionais"})
		return
	}

	var profissionais []map[string]interface{}
	for _, d := range dentistas {
		profissionais = append(profissionais, map[string]interface{}{
			"id":   d.ID,
			"nome": d.Usuario.Nome,
			"cro":  d.CRO,
		})
	}

	c.JSON(http.StatusOK, profissionais)
}
