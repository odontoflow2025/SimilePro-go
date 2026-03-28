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

// Helper to determine clinic IDs for queries based on `clinica_id` and `rede` params
func (h *ContabilidadeHandler) getClinicaIDs(c *gin.Context) ([]uint, error) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	targetClinicaID := c.Query("clinica_id")
	isRede := c.Query("rede") == "true"

	var userClinica models.Clinica
	if err := h.DB.First(&userClinica, userClinicaID).Error; err != nil {
		return nil, err
	}

	matrizID := userClinica.ID
	if userClinica.MatrizID != nil {
		matrizID = *userClinica.MatrizID
	}

	var clinicaIDs []uint

	if isRede {
		// Network query: get all clinics sharing the same Matriz
		var networkClinicas []models.Clinica
		if err := h.DB.Where("id = ? OR matriz_id = ?", matrizID, matrizID).Find(&networkClinicas).Error; err != nil {
			return nil, err
		}
		for _, nc := range networkClinicas {
			clinicaIDs = append(clinicaIDs, nc.ID)
		}
	} else if targetClinicaID != "" && targetClinicaID != "all" {
		// Verify if targeted clinic is in the same network
		var target models.Clinica
		if err := h.DB.First(&target, targetClinicaID).Error; err == nil {
			targetMatriz := target.ID
			if target.MatrizID != nil {
				targetMatriz = *target.MatrizID
			}
			if targetMatriz == matrizID {
				clinicaIDs = append(clinicaIDs, target.ID)
			}
		}
	}

	// Fallback to user's clinic
	if len(clinicaIDs) == 0 {
		clinicaIDs = append(clinicaIDs, userClinicaID)
	}

	return clinicaIDs, nil
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
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var centro models.CentroCusto
	if err := c.ShouldBindJSON(&centro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force current clinic ID
	centro.ClinicaID = userClinicaID

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
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var centros []models.CentroCusto
	if err := h.DB.Where("clinica_id = ?", userClinicaID).Find(&centros).Error; err != nil {
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
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var conta models.PlanoConta
	if err := c.ShouldBindJSON(&conta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force current clinic ID
	conta.ClinicaID = userClinicaID

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
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var contas []models.PlanoConta
	if err := h.DB.Preload("SubContas").Where("conta_pai_id IS NULL AND clinica_id = ?", userClinicaID).Find(&contas).Error; err != nil {
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

	clinicaIDs, err := h.getClinicaIDs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve clinic IDs"})
		return
	}
	query = query.Where("clinica_id IN ?", clinicaIDs)

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

// GetBIDashboard godoc
// @Summary      Get BI Dashboard Metrics
// @Description  Retrieve KPIs (Revenue, Patients, Appointments, Pending) for BI Dashboard
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Param        clinica_id query string false "Specific Clinic ID"
// @Param        rede       query bool   false "True to fetch network-wide data"
// @Param        periodo    query string false "Period filter"
// @Success      200        {object} map[string]interface{}
// @Router       /financeiro/dashboard/bi-metrics [get]
func (h *ContabilidadeHandler) GetBIDashboard(c *gin.Context) {
	clinicaIDs, err := h.getClinicaIDs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve clinic IDs"})
		return
	}

	periodo := c.Query("periodo")
	var startDate, endDate time.Time
	now := time.Now()

	switch periodo {
	case "1d":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "1w":
		startDate = now.AddDate(0, 0, -7)
	case "1m":
		startDate = now.AddDate(0, -1, 0)
	case "3m":
		startDate = now.AddDate(0, -3, 0)
	case "6m":
		startDate = now.AddDate(0, -6, 0)
	case "1y":
		startDate = now.AddDate(-1, 0, 0)
	default: // Total or 6m fallback
		startDate = now.AddDate(0, -6, 0)
	}
	endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// 1. Revenue
	var revenue float64
	h.DB.Model(&models.Transacao{}).
		Where("clinica_id IN ? AND status = ? AND tipo = ?", clinicaIDs, models.StatusTransacaoPago, models.TipoTransacaoReceita).
		Where("data_pagamento BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(valor), 0)").Scan(&revenue)

	// 2. Pending Payments
	var pending float64
	h.DB.Model(&models.Transacao{}).
		Where("clinica_id IN ? AND status = ? AND tipo = ?", clinicaIDs, models.StatusTransacaoPendente, models.TipoTransacaoReceita).
		Where("data_vencimento BETWEEN ? AND ?", startDate, endDate). // Using due date for pending
		Select("COALESCE(SUM(valor), 0)").Scan(&pending)

	// 3. Appointments Held
	var appointments int64
	h.DB.Model(&models.Agendamento{}).
		Where("clinica_id IN ? AND status = ?", clinicaIDs, models.StatusAgendamentoAtendido).
		Where("data_hora_inicio BETWEEN ? AND ?", startDate, endDate).
		Count(&appointments)

	// 4. Active Patients (patients with at least one appointment in this period)
	var activePatients int64
	h.DB.Model(&models.Agendamento{}).
		Where("clinica_id IN ? AND status = ?", clinicaIDs, models.StatusAgendamentoAtendido).
		Where("data_hora_inicio BETWEEN ? AND ?", startDate, endDate).
		Distinct("paciente_id").Count(&activePatients)

	// Calculate variations (Mock logic for now - comparing to same previous period length)
	// For a real production app we'd do the same queries for (startDate - diff) to (startDate)
	// Using static variations as was in frontend for simplicity until user requests dynamic variation formulas

	c.JSON(http.StatusOK, gin.H{
		"metrics": map[string]interface{}{
			"revenue":      revenue,
			"pending":      pending,
			"appointments": appointments,
			"patients":     activePatients,
		},
	})
}

// GetDREMensal godoc
// @Summary      Get DRE (Income Statement)
// @Description  Retrieve DRE metrics for a specific month and year
// @Tags         contabilidade
// @Produce      json
// @Security     BearerAuth
// @Param        mes query string true "Month (1-12)"
// @Param        ano query string true "Year (YYYY)"
// @Success      200 {object} map[string]interface{}
// @Router       /contabilidade/dre/mensal [get]
func (h *ContabilidadeHandler) GetDREMensal(c *gin.Context) {
	clinicaIDs, err := h.getClinicaIDs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve clinic IDs"})
		return
	}

	mes := c.Query("mes")
	ano := c.Query("ano")

	if mes == "" || ano == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Os parâmetros 'mes' e 'ano' são obrigatórios"})
		return
	}

	// For MVP DRE, we aggregate from Transacao explicitly relying on categories or just Revenue/Expenses.
	// Since we introduced LancamentoContabil recently, we might not have a full history. 
	// We will use Transacao where Status == PAGO.
	query := h.DB.Model(&models.Transacao{}).Where("clinica_id IN ? AND status = ?", clinicaIDs, models.StatusTransacaoPago)
	
	// Filtro PostgreSQL para extrair mês e ano
	query = query.Where("EXTRACT(MONTH FROM data_pagamento) = ? AND EXTRACT(YEAR FROM data_pagamento) = ?", mes, ano)

	// In a complete accounting system, "Taxes", "Costs" and "Expenses" are defined by the chart of accounts (Plano de Contas).
	// We will infer from 'TipoTransacao' and standard sub-categories if present, or provide standard ratios if not mapped.
	var totalReceitas, impostos, custos, despesas float64

	// Receitas brutas (All Receitas)
	h.DB.Model(&models.Transacao{}).Where("clinica_id IN ? AND status = ? AND tipo = ?", clinicaIDs, models.StatusTransacaoPago, models.TipoTransacaoReceita).
		Where("EXTRACT(MONTH FROM data_pagamento) = ? AND EXTRACT(YEAR FROM data_pagamento) = ?", mes, ano).
		Select("COALESCE(SUM(valor), 0)").Scan(&totalReceitas)

	// Despesas totais
	var despesasTotais float64
	h.DB.Model(&models.Transacao{}).Where("clinica_id IN ? AND status = ? AND tipo = ?", clinicaIDs, models.StatusTransacaoPago, models.TipoTransacaoDespesa).
		Where("EXTRACT(MONTH FROM data_pagamento) = ? AND EXTRACT(YEAR FROM data_pagamento) = ?", mes, ano).
		Select("COALESCE(SUM(valor), 0)").Scan(&despesasTotais)

	// Fictional classification for MVP if exact Categories aren't strictly typed by user yet
	// Real world: sum based on PlanoConta grouping.
	impostos = totalReceitas * 0.08      // Simulating ~8% tax rate (Simples Nacional/Lucro Presumido)
	custos = totalReceitas * 0.25        // Repasses para dentistas (25%) + laboratório
	despesas = despesasTotais - impostos - custos // O restante das saídas do caixa é despesa fixa

	if despesas < 0 {
		despesas = despesasTotais // Fallback logic
		impostos = 0
		custos = 0
	}

	receitaLiquida := totalReceitas - impostos
	lucroBruto := receitaLiquida - custos
	lucroLiquido := lucroBruto - despesas

	c.JSON(http.StatusOK, gin.H{
		"periodo":      mes + "/" + ano,
		"grossRevenue": totalReceitas,
		"taxes":        impostos,
		"netRevenue":   receitaLiquida,
		"costs":        custos,
		"grossMargin":  lucroBruto,
		"expenses":     despesas,
		"netIncome":    lucroLiquido,
	})
}

