package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssinaturaHandler struct {
	DB      *gorm.DB
	Service *services.AssinaturaService
}

func NewAssinaturaHandler(db *gorm.DB) *AssinaturaHandler {
	return &AssinaturaHandler{
		DB:      db,
		Service: services.NewAssinaturaService(db),
	}
}

// GetPlanos returns available plans
// GetPlanos godoc
// @Summary      Available plans
// @Description  List all SaaS subscription plans and their features
// @Tags         assinaturas
// @Produce      json
// @Success      200  {array}  object
// @Router       /assinatura/planos [get]
func (h *AssinaturaHandler) GetPlanos(c *gin.Context) {
	planos := []gin.H{
		{
			"id":          "FREE",
			"nome":        "Plano Gratuito",
			"preco":       0.00,
			"funcionarios": 2,
			"features":    []string{"Agendamento Básico", "Cadastro de Pacientes"},
		},
		{
			"id":          "BASE",
			"nome":        "Plano Básico",
			"preco":       99.90,
			"funcionarios": 5,
			"features":    []string{"Agendamento Completo", "Prontuário Eletrônico", "Financeiro Básico"},
		},
		{
			"id":          "PREMIUM",
			"nome":        "Plano Premium",
			"preco":       199.90,
			"funcionarios": 9999,
			"features":    []string{"Tudo do Básico", "Dashboard Avançado", "RNDS", "Múltiplas Unidades"},
		},
		{
			"id":          "EDUCACIONAL",
			"nome":        "Plano Educacional",
			"preco":       0.00,
			"funcionarios": 9999,
			"features":    []string{"Uso Acadêmico", "Sem Financeiro"},
		},
	}
	c.JSON(http.StatusOK, planos)
}

// GetStatus returns current clinic status and limits
// GetStatus godoc
// @Summary      Subscription status
// @Description  Get current plan and status for the user's clinic
// @Tags         assinaturas
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /assinatura/status [get]
func (h *AssinaturaHandler) GetStatus(c *gin.Context) {
	clinicaIDVal, _ := c.Get("clinicaID")
	if clinicaIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	var clinicaID uint
	switch v := clinicaIDVal.(type) {
	case float64:
		clinicaID = uint(v)
	case uint:
		clinicaID = v
	case int:
		clinicaID = uint(v)
	}

	var clinica models.Clinica
	if err := h.DB.First(&clinica, clinicaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinica not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clinicaId":       clinica.ID,
		"plano":           clinica.Plano,
		"status":          clinica.PlanoStatus,
		"planoExpiracao":  clinica.PlanoExpiracao,
		"isActive":        h.Service.IsActive(&clinica),
	})
}

// Assinar updates the plan
// Assinar godoc
// @Summary      Change plan
// @Description  Update the clinic's subscription plan
// @Tags         assinaturas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      object  true  "New Plan ID and Periodivity"
// @Success      200    {object}  map[string]string
// @Router       /assinatura/assinar [post]
func (h *AssinaturaHandler) Assinar(c *gin.Context) {
	userClinicaIDVal, _ := c.Get("clinicaID")
	if userClinicaIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	var input struct {
		Plano         string `json:"plano" binding:"required"`
		Periodicidade string `json:"periodicidade"` // MENSAL or ANUAL
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    // Default to MENSAL if not provided
    if input.Periodicidade == "" {
        input.Periodicidade = "MENSAL"
    }

    var clinicaID uint
	switch v := userClinicaIDVal.(type) {
	case float64:
		clinicaID = uint(v)
	case uint:
		clinicaID = v
	case int:
		clinicaID = uint(v)
	}

	if err := h.Service.UpdatePlan(clinicaID, input.Plano, input.Periodicidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plano atualizado com sucesso"})
}
