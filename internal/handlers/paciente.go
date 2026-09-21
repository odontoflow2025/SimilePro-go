package handlers

import (
	"fmt"
	"net/http"
	"SimilePro-go/internal/models"
	"SimilePro-go/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PacienteHandler struct {
    DB *gorm.DB
}

func NewPacienteHandler(db *gorm.DB) *PacienteHandler {
    return &PacienteHandler{DB: db}
}

type CreatePacienteInput struct {
    Nome              string    `json:"nome" binding:"required"`
    CPF               string    `json:"cpf"`
    RG                string    `json:"rg"`
    DataNascimento    time.Time `json:"dataNascimento"`
    Genero            string    `json:"genero"`
    TelefonePrincipal string    `json:"telefonePrincipal" binding:"required"`
    Email             string    `json:"email"`
    EnderecoCompleto  string    `json:"enderecoCompleto"`
    ClinicaID         uint      `json:"clinicaId" binding:"required"`
}

// Create godoc
// @Summary      Create a new patient
// @Description  Create a new patient record in the system
// @Tags         pacientes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreatePacienteInput  true  "Patient Info"
// @Success      201    {object}  models.Paciente
// @Failure      400    {object}  map[string]string
// @Router       /pacientes [post]
func (h *PacienteHandler) Create(c *gin.Context) {
	var input CreatePacienteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force ClinicaID from Authenticated User context for data isolation
	userClinicaID, exists := c.Get("clinicaID")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not associated with any clinic"})
		return
	}

	clinicaID := userClinicaID.(uint)

	// Generate unique code
	codigoUnico := "OF-" + time.Now().Format("2006") + "-" + strconv.FormatInt(time.Now().UnixNano()%10000, 10)

	paciente := models.Paciente{
		Nome:              input.Nome,
		CPF:               input.CPF,
		CodigoUnico:       codigoUnico,
		RG:                input.RG,
		DataNascimento:    input.DataNascimento,
		Genero:            input.Genero,
		TelefonePrincipal: input.TelefonePrincipal,
		Email:             input.Email,
		EnderecoCompleto:  input.EnderecoCompleto,
		ClinicaID:         clinicaID,
	}

	if err := h.DB.Create(&paciente).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create paciente"})
		return
	}

	c.JSON(http.StatusCreated, paciente)
}

// FindAll godoc
// @Summary      List all patients
// @Description  List patients with global search by CPF/Code and clinic filtering
// @Tags         pacientes
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Param        search     query     string  false  "Global search by CPF or Code"
// @Success      200        {array}   models.Paciente
// @Router       /pacientes [get]
func (h *PacienteHandler) FindAll(c *gin.Context) {
	search := c.Query("search") // Can be CPF or CodigoUnico

	tenantDBRaw, exists := c.Get("tenantDB")
	var query *gorm.DB
	if exists {
		query = tenantDBRaw.(*gorm.DB)
	} else {
		userClinicaID, _ := c.Get("clinicaID")
		query = h.DB.Where("clinica_id = ?", userClinicaID)
	}

	var pacientes []models.Paciente

	if search != "" {
		// Converter a pesquisa em Hash se for CPF
		searchHash, _ := utils.HashDeterministic(search)
		query = query.Where("cpf_hash = ? OR codigo_unico = ?", searchHash, search)
	}

	if err := query.Find(&pacientes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pacientes"})
		return
	}

	c.JSON(http.StatusOK, pacientes)
}

// FindOne godoc
// @Summary      Get patient details
// @Description  Get full details of a patient. Restricts by TenantDB.
// @Tags         pacientes
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true   "Patient ID"
// @Success      200        {object}  models.Paciente
// @Failure      404        {object}  map[string]string
// @Router       /pacientes/{id} [get]
func (h *PacienteHandler) FindOne(c *gin.Context) {
	id := c.Param("id")
	
	tenantDBRaw, exists := c.Get("tenantDB")
	var query *gorm.DB
	if exists {
		query = tenantDBRaw.(*gorm.DB)
	} else {
		userClinicaID, _ := c.Get("clinicaID")
		query = h.DB.Where("clinica_id = ?", userClinicaID)
	}

	var paciente models.Paciente
	if err := query.Preload("Clinica").First(&paciente, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	c.JSON(http.StatusOK, paciente)
}

// GerarProtocolo godoc
// @Summary      Generate access protocol
// @Description  Generate a temporary audit protocol to access sensitive patient data
// @Tags         pacientes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      object  true  "Protocol Request"
// @Success      201    {object}  map[string]string
// @Router       /pacientes/protocolo-acesso [post]
func (h *PacienteHandler) GerarProtocolo(c *gin.Context) {
	var input struct {
		PacienteID    uint   `json:"pacienteId" binding:"required"`
		Justificativa string `json:"justificativa" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// Generate a random 8-char protocol
	protocolNum := fmt.Sprintf("PRT-%d-%X", input.PacienteID, time.Now().UnixNano()%1000000)

	acesso := models.AcessoProntuario{
		PacienteID:      input.PacienteID,
		UsuarioID:       userID.(uint),
		NumeroProtocolo: protocolNum,
		Justificativa:   input.Justificativa,
		Expiracao:       time.Now().Add(time.Hour * 4), // Valid for 4 hours
	}

	if err := h.DB.Create(&acesso).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar protocolo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"numeroProtocolo": protocolNum,
		"expiracao":       acesso.Expiracao,
	})
}
