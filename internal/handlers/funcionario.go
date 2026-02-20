package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
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
	var input CreateFuncionarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	funcionario := models.Funcionario{
		UsuarioID:    input.UsuarioID,
		ClinicaID:    input.ClinicaID,
		Cargo:        input.Cargo,
		DataAdmissao: input.DataAdmissao,
		Salario:      input.Salario,
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
	clinicaID := c.Query("clinicaId")

	var funcionarios []models.Funcionario
	query := h.DB.Preload("Usuario").Preload("Clinica")

	if clinicaID != "" {
		query = query.Where("clinica_id = ?", clinicaID)
	}

	if err := query.Find(&funcionarios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch funcionarios"})
		return
	}

	c.JSON(http.StatusOK, funcionarios)
}
