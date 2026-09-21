package handlers

import (
	"net/http"
	"SimilePro-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FuncionarioHandler struct {
    DB *gorm.DB
}

func NewFuncionarioHandler(db *gorm.DB) *FuncionarioHandler {
    return &FuncionarioHandler{DB: db}
}

type CreateFuncionarioInput struct {
    UsuarioID    uint      `json:"usuarioId" binding:"required"`
    ClinicaID    uint      `json:"clinicaId" binding:"required"`
    Cargo        string    `json:"cargo" binding:"required"`
    DataAdmissao time.Time `json:"dataAdmissao"`
    Salario      float64   `json:"salario"`
    PIS          string    `json:"pis"`
    CTPS         string    `json:"ctps"`
    RG           string    `json:"rg"`
    CNPJ         string    `json:"cnpj"`
    Endereco     string    `json:"endereco"`
    DadosBanco   string    `json:"dadosBanco"`
    CargaHoraria int       `json:"cargaHoraria"`
    Dependentes  int       `json:"dependentes"`
}

// Create godoc
// @Summary      Register an employee
// @Description  Register a new administrative or health support employee
// @Tags         funcionarios
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateFuncionarioInput  true  "Employee Info"
// @Success      201    {object}  models.Funcionario
// @Router       /funcionarios [post]
func (h *FuncionarioHandler) Create(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input CreateFuncionarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	funcionario := models.Funcionario{
		UsuarioID:    input.UsuarioID,
		ClinicaID:    userClinicaID, // Force current clinic ID
		Cargo:        input.Cargo,
		DataAdmissao: input.DataAdmissao,
		Salario:      input.Salario,
        PIS:          input.PIS,
        CTPS:         input.CTPS,
        RG:           input.RG,
        CNPJ:         input.CNPJ,
        Endereco:     input.Endereco,
        DadosBanco:   input.DadosBanco,
        CargaHoraria: input.CargaHoraria,
        Dependentes:  input.Dependentes,
	}

	if err := h.DB.Create(&funcionario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create funcionario"})
		return
	}

	c.JSON(http.StatusCreated, funcionario)
}

// FindAll godoc
// @Summary      List employees
// @Description  Retrieve a list of all employees (with clinic filtering)
// @Tags         funcionarios
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Success      200        {array}   models.Funcionario
// @Router       /funcionarios [get]
func (h *FuncionarioHandler) FindAll(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	reqClinicaID := c.Query("clinicaId")
	query := h.DB.Preload("Usuario").Preload("Clinica")

	if userRole == "ADMIN_TOTAL" {
		if reqClinicaID != "" {
			query = query.Where("clinica_id = ?", reqClinicaID)
		}
	} else {
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	var funcionarios []models.Funcionario
	if err := query.Find(&funcionarios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch funcionarios"})
		return
	}

	c.JSON(http.StatusOK, funcionarios)
}

type DemitirFuncionarioInput struct {
    DataDemissao   time.Time `json:"dataDemissao" binding:"required"`
    MotivoDemissao string    `json:"motivoDemissao" binding:"required"`
}

// Demitir godoc
// @Summary      Dismiss an employee
// @Description  Change employee status to DEMITIDO and record reason/date
// @Tags         funcionarios
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string                   true  "Employee ID"
// @Param        input  body      DemitirFuncionarioInput  true  "Dismissal Info"
// @Success      200    {object}  models.Funcionario
// @Router       /funcionarios/{id}/demitir [patch]
func (h *FuncionarioHandler) Demitir(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	var input DemitirFuncionarioInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var funcionario models.Funcionario
	if err := h.DB.First(&funcionario, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Funcionario não encontrado"})
		return
	}

	if userRole != "ADMIN_TOTAL" && funcionario.ClinicaID != userClinicaID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Não autorizado a demitir funcionário de outra clínica"})
		return
	}

    if funcionario.Status == models.StatusFuncionarioDemitido {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Funcionário já está demitido"})
		return
    }

	funcionario.Status = models.StatusFuncionarioDemitido
	funcionario.DataDemissao = &input.DataDemissao
	funcionario.MotivoDemissao = input.MotivoDemissao

	if err := h.DB.Save(&funcionario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao registrar rescisão"})
		return
	}

	c.JSON(http.StatusOK, funcionario)
}
