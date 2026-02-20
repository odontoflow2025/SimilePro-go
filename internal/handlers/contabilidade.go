package handlers

import (
	"log"
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContabilidadeHandler struct {
    DB *gorm.DB
}

func NewContabilidadeHandler(db *gorm.DB) *ContabilidadeHandler {
    return &ContabilidadeHandler{DB: db}
}

// --- Centro de Custo ---

// CreateCentroCusto godoc
// @Summary      Create cost center
// @Description  Create a new cost center for organizational accounting
// @Tags         contabilidade
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      models.CentroCusto  true  "Cost Center Info"
// @Success      201    {object}  models.CentroCusto
// @Router       /financeiro/centros-custo [post]
func (h *ContabilidadeHandler) CreateCentroCusto(c *gin.Context) {
	var centro models.CentroCusto
	if err := c.ShouldBindJSON(&centro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&centro).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create centro de custo"})
		return
	}

	c.JSON(http.StatusCreated, centro)
}

// FindAllCentrosCusto godoc
// @Summary      List cost centers
// @Description  Retrieve all active cost centers for the clinic
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.CentroCusto
// @Router       /financeiro/centros-custo [get]
func (h *ContabilidadeHandler) FindAllCentrosCusto(c *gin.Context) {
	var centros []models.CentroCusto
	if err := h.DB.Find(&centros).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch centros de custo"})
		return
	}

	c.JSON(http.StatusOK, centros)
}

// --- Plano de Contas ---

// CreateConta godoc
// @Summary      Create account plan item
// @Description  Add a new category/account to the chart of accounts
// @Tags         contabilidade
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      models.PlanoConta  true  "Account Info"
// @Success      201    {object}  models.PlanoConta
// @Router       /financeiro/plano-contas [post]
func (h *ContabilidadeHandler) CreateConta(c *gin.Context) {
	var conta models.PlanoConta
	if err := c.ShouldBindJSON(&conta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&conta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conta"})
		return
	}

	c.JSON(http.StatusCreated, conta)
}

// FindAllContas godoc
// @Summary      Get chart of accounts
// @Description  Retrieve the full hierarchical chart of accounts
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.PlanoConta
// @Router       /financeiro/plano-contas [get]
func (h *ContabilidadeHandler) FindAllContas(c *gin.Context) {
	var contas []models.PlanoConta
	if err := h.DB.Preload("SubContas").Where("conta_pai_id IS NULL").Find(&contas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plano de contas"})
		return
	}

	c.JSON(http.StatusOK, contas)
}

// GetEstrutura godoc
// @Summary      Get account hierarchy
// @Description  Retrieve the full chart of accounts in a recursive tree structure
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.PlanoConta
// @Router       /financeiro/plano-contas/estrutura [get]
func (h *ContabilidadeHandler) GetEstrutura(c *gin.Context) {
    // Similar to FindAll but potentially more recursive or formatted tree
    h.FindAllContas(c)
}


// GetFluxoCaixa godoc
// @Summary      Get cash flow data
// @Description  Retrieve data points for cash flow chart based on period (e.g., 7DIAS, 30DIAS)
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Param        periodo  query     string  false  "Period (7DIAS, 30DIAS, TRIMESTRE, SEMESTRE, ANO, TOTAL, HOJE)"
// @Success      200      {object}  map[string]interface{}
// @Router       /financeiro/fluxo-caixa [get]
func (h *ContabilidadeHandler) GetFluxoCaixa(c *gin.Context) {
	periodo := c.Query("periodo")
	var startDate, endDate time.Time
	now := time.Now()

	// Determine date range
	switch periodo {
	case "7DIAS":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "30DIAS":
		startDate = now.AddDate(0, 0, -30)
		endDate = now
	case "TRIMESTRE":
		startDate = now.AddDate(0, -3, 0)
		endDate = now
	case "SEMESTRE":
		startDate = now.AddDate(0, -6, 0)
		endDate = now
	case "ANO":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "TOTAL":
		startDate = time.Time{} // Zero value
		endDate = now
	case "HOJE":
		startDate = now
		endDate = now
	default:
		startDate = now.AddDate(0, 0, -30)
		endDate = now
	}

	// Normalize times
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	if !startDate.IsZero() {
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	}

	// struct to hold query results
	var results []struct {
		DateStr string  `gorm:"column:date_str"`
		Tipo    string  `gorm:"column:tipo"`
		Total   float64 `gorm:"column:total"`
	}

	// Build Query with Debug logging
	query := h.DB.Debug().Model(&models.Transacao{}).
		Select("TO_CHAR(data_vencimento, 'YYYY-MM-DD') as date_str, tipo, SUM(valor) as total").
		Where("status != ?", models.StatusTransacaoCancelado)

	if !startDate.IsZero() {
		query = query.Where("data_vencimento BETWEEN ? AND ?", startDate, endDate)
	} else {
		query = query.Where("data_vencimento <= ?", endDate)
	}

	// Group and Order
	if err := query.Group("date_str, tipo").Order("date_str ASC").Scan(&results).Error; err != nil {
		log.Printf("Error querying cash flow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chart data"})
		return
	}

	log.Printf("Cash Flow Query found %d rows for period %s", len(results), periodo)

	// Process results into expected format
	type DataPoint struct {
		Data     string  `json:"data"`
		Entradas float64 `json:"entradas"`
		Saidas   float64 `json:"saidas"`
	}

	var fluxo []DataPoint
	if len(results) > 0 {
		currentDate := ""
		var currentPoint *DataPoint

		for _, r := range results {
			if r.DateStr != currentDate {
				if currentPoint != nil {
					fluxo = append(fluxo, *currentPoint)
				}
				currentDate = r.DateStr
				currentPoint = &DataPoint{Data: r.DateStr, Entradas: 0, Saidas: 0}
			}
			if r.Tipo == string(models.TipoTransacaoReceita) {
				currentPoint.Entradas = r.Total
			} else if r.Tipo == string(models.TipoTransacaoDespesa) {
				currentPoint.Saidas = r.Total
			}
		}
		if currentPoint != nil {
			fluxo = append(fluxo, *currentPoint)
		}
	} else {
		fluxo = []DataPoint{}
	}

	c.JSON(http.StatusOK, gin.H{"fluxo": fluxo})
}
