package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransacaoHandler struct {
    DB *gorm.DB
}

func NewTransacaoHandler(db *gorm.DB) *TransacaoHandler {
    return &TransacaoHandler{DB: db}
}

type CreateTransacaoInput struct {
    ClinicaID      uint                   `json:"clinicaId" binding:"required"`
    PacienteID     *uint                  `json:"pacienteId"`
    Descricao      string                 `json:"descricao" binding:"required"`
    Valor          float64                `json:"valor" binding:"required"`
    Tipo           models.TipoTransacao   `json:"tipo" binding:"required"`
    DataVencimento time.Time              `json:"dataVencimento" binding:"required"`
    Categoria      string                 `json:"categoria"`
    FormaPagamento string                 `json:"formaPagamento"`
}

type UpdateStatusTransacaoInput struct {
    Status        models.StatusTransacao `json:"status" binding:"required"`
    DataPagamento *time.Time             `json:"dataPagamento"`
}

// Create godoc
// @Summary      Register a transaction
// @Description  Create a new financial transaction (Revenue/Expense) for the clinic
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateTransacaoInput  true  "Transaction Info"
// @Success      201    {object}  models.Transacao
// @Router       /financeiro/transacoes [post]
func (h *TransacaoHandler) Create(c *gin.Context) {
	var input CreateTransacaoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transacao := models.Transacao{
		ClinicaID:      input.ClinicaID,
		PacienteID:     input.PacienteID,
		Descricao:      input.Descricao,
		Valor:          input.Valor,
		Tipo:           input.Tipo,
		Status:         models.StatusTransacaoPendente,
		DataVencimento: input.DataVencimento,
		Categoria:      input.Categoria,
		FormaPagamento: input.FormaPagamento,
	}

	if err := h.DB.Create(&transacao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transacao"})
		return
	}

	c.JSON(http.StatusCreated, transacao)
}

// FindAll godoc
// @Summary      List transactions
// @Description  Retrieve a filtered list of financial transactions
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Param        tipo       query     string  false  "Filter by Type (RECEITA/DESPESA)"
// @Param        status     query     string  false  "Filter by Status"
// @Param        dataInicio query     string  false  "Start Date (YYYY-MM-DD)"
// @Param        dataFim    query     string  false  "End Date (YYYY-MM-DD)"
// @Success      200        {array}   models.Transacao
// @Router       /financeiro/transacoes [get]
func (h *TransacaoHandler) FindAll(c *gin.Context) {
	clinicaID := c.Query("clinicaId")
	tipo := c.Query("tipo")
	status := c.Query("status")
	dataInicio := c.Query("dataInicio")
	dataFim := c.Query("dataFim")

	var transacoes []models.Transacao
	query := h.DB.Preload("Paciente")

	if clinicaID != "" {
		query = query.Where("clinica_id = ?", clinicaID)
	}
	if tipo != "" {
		query = query.Where("tipo = ?", tipo)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if dataInicio != "" {
		query = query.Where("data_vencimento >= ?", dataInicio)
	}
	if dataFim != "" {
		query = query.Where("data_vencimento <= ?", dataFim)
	}

	if err := query.Order("data_vencimento asc").Find(&transacoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transacoes"})
		return
	}

	c.JSON(http.StatusOK, transacoes)
}

// UpdateStatus godoc
// @Summary      Update transaction status
// @Description  Mark a transaction as paid, overdue, or cancelled
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string                      true  "Transaction ID"
// @Param        input  body      UpdateStatusTransacaoInput  true  "New Status"
// @Success      200    {object}  models.Transacao
// @Router       /financeiro/transacoes/{id}/status [patch]
func (h *TransacaoHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var input UpdateStatusTransacaoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var transacao models.Transacao
	if err := h.DB.First(&transacao, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transacao not found"})
		return
	}

	transacao.Status = input.Status
	if input.DataPagamento != nil {
		transacao.DataPagamento = input.DataPagamento
	} else if input.Status == models.StatusTransacaoPago && transacao.DataPagamento == nil {
		now := time.Now()
		transacao.DataPagamento = &now
	}

	if err := h.DB.Save(&transacao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, transacao)
}


// GetSummary godoc
// @Summary      Cash balance summary
// @Description  Get totals for revenues, expenses, and net balance for a period
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Param        dataInicio query     string  false  "Start Date (YYYY-MM-DD)"
// @Param        dataFim    query     string  false  "End Date (YYYY-MM-DD)"
// @Success      200        {object}  map[string]float64
// @Router       /financeiro/sumario [get]
func (h *TransacaoHandler) GetSummary(c *gin.Context) {
	dataInicio := c.Query("dataInicio")
	dataFim := c.Query("dataFim")

	// Helper to build filtering query
	buildQuery := func(tipo models.TipoTransacao) *gorm.DB {
		q := h.DB.Model(&models.Transacao{}).Where("tipo = ? AND status != ?", tipo, models.StatusTransacaoCancelado)
		if dataInicio != "" {
			q = q.Where("data_vencimento >= ?", dataInicio)
		}
		if dataFim != "" {
			// Add time to end date to include the whole day if it's just a date string
			if len(dataFim) == 10 { // YYYY-MM-DD
				q = q.Where("data_vencimento <= ?", dataFim+" 23:59:59")
			} else {
				q = q.Where("data_vencimento <= ?", dataFim)
			}
		}
		return q
	}

	type Result struct {
		Total float64
	}

	var entradas Result
	buildQuery(models.TipoTransacaoReceita).Select("COALESCE(SUM(valor), 0) as total").Scan(&entradas)

	var saidas Result
	buildQuery(models.TipoTransacaoDespesa).Select("COALESCE(SUM(valor), 0) as total").Scan(&saidas)

	summary := gin.H{
		"totalEntradas": entradas.Total,
		"totalSaidas":   saidas.Total,
		"saldo":         entradas.Total - saidas.Total,
		"receitas":      entradas.Total,
		"despesas":      saidas.Total,
		"resultado":     entradas.Total - saidas.Total,
	}

	c.JSON(http.StatusOK, summary)
}
