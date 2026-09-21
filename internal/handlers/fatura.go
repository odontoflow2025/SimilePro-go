package handlers

import (
	"net/http"
	"SimilePro-go/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FaturaHandler struct {
    DB *gorm.DB
}

func NewFaturaHandler(db *gorm.DB) *FaturaHandler {
    return &FaturaHandler{DB: db}
}

type CreateFaturaInput struct {
    ClinicaID         uint      `json:"clinicaId" binding:"required"`
    PacienteID        uint      `json:"pacienteId" binding:"required"`
    PlanoTratamentoID *uint     `json:"planoTratamentoId"`
    ValorTotal        float64   `json:"valorTotal" binding:"required,gt=0,lt=1000000000"`
    DataVencimento    time.Time `json:"dataVencimento" binding:"required"`
}

type PagarFaturaInput struct {
    FormaPagamento string `json:"formaPagamento" binding:"required"`
}

// Create godoc
// @Summary      Create an invoice
// @Description  Generate a new invoice for a patient, optionally linked to a treatment plan
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateFaturaInput  true  "Invoice Info"
// @Success      201    {object}  models.Fatura
// @Router       /financeiro/faturas [post]
func (h *FaturaHandler) Create(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input CreateFaturaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fatura := models.Fatura{
		ClinicaID:         userClinicaID, // Force current clinic ID
		PacienteID:        input.PacienteID,
		PlanoTratamentoID: input.PlanoTratamentoID,
		ValorTotal:        input.ValorTotal,
		Status:            models.StatusFaturaPendente,
		DataGeracao:       time.Now(),
		DataVencimento:    input.DataVencimento,
	}

	if err := h.DB.Create(&fatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fatura"})
		return
	}

	c.JSON(http.StatusCreated, fatura)
}

type UpdateFaturaInput struct {
    ValorTotal        float64   `json:"valorTotal" binding:"required,gt=0,lt=1000000000"`
    DataVencimento    time.Time `json:"dataVencimento" binding:"required"`
}

// Update godoc
// @Summary      Update an invoice
// @Description  Update details of an open invoice. Paid invoices cannot be updated.
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Invoice ID"
// @Param        input  body      UpdateFaturaInput  true  "Invoice Info"
// @Success      200    {object}  models.Fatura
// @Router       /financeiro/faturas/{id} [put]
func (h *FaturaHandler) Update(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	var input UpdateFaturaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fatura models.Fatura
	if err := h.DB.First(&fatura, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fatura não encontrada"})
		return
	}

	// Ownership check (BOLA)
	if userRole != "ADMIN_TOTAL" && fatura.ClinicaID != userClinicaID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Não autorizado a alterar fatura de outra clínica"})
		return
	}

	// Proteção contra Mutabilidade de Faturas Pagas
	if fatura.Status == models.StatusFaturaPaga || fatura.Status == models.StatusFaturaCancelada {
		c.JSON(http.StatusConflict, gin.H{"error": "Faturas pagas ou canceladas não podem ser alteradas"})
		return
	}

	fatura.ValorTotal = input.ValorTotal
	fatura.DataVencimento = input.DataVencimento

	if err := h.DB.Save(&fatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar fatura"})
		return
	}

	c.JSON(http.StatusOK, fatura)
}


// FindAll godoc
// @Summary      List invoices
// @Description  Retrieve a list of invoices with filtering by clinic, patient and status
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Param        clinicaId  query     string  false  "Filter by Clinic ID"
// @Param        pacienteId query     string  false  "Filter by Patient ID"
// @Param        status     query     string  false  "Filter by Status"
// @Success      200        {array}   models.Fatura
// @Router       /financeiro/faturas [get]
func (h *FaturaHandler) FindAll(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	reqClinicaID := c.Query("clinicaId")
	pacienteID := c.Query("pacienteId")
	status := c.Query("status")

	var faturas []models.Fatura
	query := h.DB.Preload("Paciente").Preload("PlanoTratamento")

	if userRole == "ADMIN_TOTAL" && reqClinicaID != "" {
		query = query.Where("clinica_id = ?", reqClinicaID)
	} else {
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	if pacienteID != "" {
		query = query.Where("paciente_id = ?", pacienteID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&faturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch faturas"})
		return
	}

	c.JSON(http.StatusOK, faturas)
}


// GetAtrasadas godoc
// @Summary      List overdue invoices
// @Description  Retrieve all invoices that are past their due date and remain unpaid
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Fatura
// @Router       /financeiro/faturas/atrasadas [get]
func (h *FaturaHandler) GetAtrasadas(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var faturas []models.Fatura
	now := time.Now()

	if err := h.DB.Preload("Paciente").Preload("PlanoTratamento").
		Where("clinica_id = ? AND status = ? AND data_vencimento < ?", userClinicaID, models.StatusFaturaPendente, now).
		Find(&faturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overdue invoices"})
		return
	}

	c.JSON(http.StatusOK, faturas)
}

// PagarFatura godoc
// @Summary      Mark an invoice as Paid
// @Description  Pays an invoice and automatically generates a Financial Revenue Transaction
// @Tags         financeiro
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Invoice ID"
// @Param        input  body      PagarFaturaInput   true  "Payment Method Info"
// @Success      200    {object}  models.Fatura
// @Router       /financeiro/faturas/{id}/pagar [post]
func (h *FaturaHandler) PagarFatura(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	var input PagarFaturaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Forma de pagamento é obrigatória"})
		return
	}

	var fatura models.Fatura
	if err := h.DB.First(&fatura, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fatura não encontrada"})
		return
	}

	// Ownership check
	if userRole != "ADMIN_TOTAL" && fatura.ClinicaID != userClinicaID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Não autorizado a pagar fatura de outra clínica"})
		return
	}

	if fatura.Status != models.StatusFaturaPendente {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A fatura não está pendente"})
		return
	}

	// Começar transação GORM para garantir integridade
	tx := h.DB.Begin()

	fatura.Status = models.StatusFaturaPaga
	if err := tx.Save(&fatura).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar fatura"})
		return
	}

	// Criar Transacao Financeira (Receita Realizada do Caixa)
	now := time.Now()
	transacao := models.Transacao{
		ClinicaID:      fatura.ClinicaID,
		PacienteID:     &fatura.PacienteID,
		Descricao:      "Pagamento Fatura ID " + strconv.Itoa(int(fatura.ID)),
		Valor:          fatura.ValorTotal,
		Tipo:           models.TipoTransacaoReceita,
		Status:         models.StatusTransacaoPago,
		DataVencimento: fatura.DataVencimento,
		DataPagamento:  &now,
		Categoria:      "Tratamentos Clínicos",
		FormaPagamento: input.FormaPagamento,
	}

	if err := tx.Create(&transacao).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar transação de receita"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, fatura)
}

// CancelarFatura godoc
// @Summary      Cancel an invoice
// @Description  Marks an open invoice as cancelled
// @Tags         financeiro
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string  true  "Invoice ID"
// @Success      200    {object}  models.Fatura
// @Router       /financeiro/faturas/{id}/cancelar [post]
func (h *FaturaHandler) CancelarFatura(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	
	var fatura models.Fatura
	if err := h.DB.First(&fatura, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fatura não encontrada"})
		return
	}

	// Ownership check
	if userRole != "ADMIN_TOTAL" && fatura.ClinicaID != userClinicaID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Não autorizado a cancelar fatura de outra clínica"})
		return
	}

	if fatura.Status == models.StatusFaturaPaga {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fatura paga não pode ser cancelada"})
		return
	}

	fatura.Status = models.StatusFaturaCancelada
	if err := h.DB.Save(&fatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao cancelar fatura"})
		return
	}

	c.JSON(http.StatusOK, fatura)
}
