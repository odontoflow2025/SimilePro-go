package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClinicaHandler struct {
    DB *gorm.DB
}

func NewClinicaHandler(db *gorm.DB) *ClinicaHandler {
    return &ClinicaHandler{DB: db}
}

type CreateClinicaInput struct {
    NomeFantasia  string `json:"nomeFantasia" binding:"required"`
    RazaoSocial   string `json:"razaoSocial" binding:"required"`
    CNPJ          string `json:"cnpj" binding:"required"`
    Endereco      string `json:"endereco"`
    Telefone      string `json:"telefone"`
    ResponsavelID uint   `json:"responsavelId"`
}

// Create godoc
// @Summary      Create a new clinic
// @Description  Register a new clinic in the system
// @Tags         clinicas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateClinicaInput  true  "Clinic Info"
// @Success      201    {object}  models.Clinica
// @Router       /clinicas [post]
func (h *ClinicaHandler) Create(c *gin.Context) {
	var input CreateClinicaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clinica := models.Clinica{
		NomeFantasia:  input.NomeFantasia,
		RazaoSocial:   input.RazaoSocial,
		CNPJ:          input.CNPJ,
		Endereco:      input.Endereco,
		Telefone:      input.Telefone,
		ResponsavelID: input.ResponsavelID,
	}

	if err := h.DB.Create(&clinica).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinica"})
		return
	}

	c.JSON(http.StatusCreated, clinica)
}

// FindAll godoc
// @Summary      List all clinics
// @Description  Retrieve a list of all registered clinics
// @Tags         clinicas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Clinica
// @Router       /clinicas [get]
func (h *ClinicaHandler) FindAll(c *gin.Context) {
	var clinicas []models.Clinica
	if err := h.DB.Find(&clinicas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinicas"})
		return
	}

	c.JSON(http.StatusOK, clinicas)
}

// FindOne godoc
// @Summary      Get clinic details
// @Description  Retrieve detailed information about a specific clinic
// @Tags         clinicas
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Clinic ID"
// @Success      200  {object}  models.Clinica
// @Failure      404  {object}  map[string]string
// @Router       /clinicas/{id} [get]
func (h *ClinicaHandler) FindOne(c *gin.Context) {
	id := c.Param("id")
	var clinica models.Clinica
	if err := h.DB.First(&clinica, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinica not found"})
		return
	}

	c.JSON(http.StatusOK, clinica)
}

// GetRede godoc
// @Summary      Get clinic network
// @Description  Retrieve all clinics that belong to the same network (Matriz and Filiais)
// @Tags         clinicas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Clinica
// @Router       /clinicas/rede [get]
func (h *ClinicaHandler) GetRede(c *gin.Context) {
	userClinicaID, _ := c.Get("clinicaID")
	if userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	var userClinica models.Clinica
	if err := h.DB.First(&userClinica, userClinicaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User clinic not found"})
		return
	}

	matrizID := userClinica.ID
	if userClinica.MatrizID != nil {
		matrizID = *userClinica.MatrizID
	}

	var clinicas []models.Clinica
	if err := h.DB.Where("id = ? OR matriz_id = ?", matrizID, matrizID).Find(&clinicas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinic network"})
		return
	}

	c.JSON(http.StatusOK, clinicas)
}
