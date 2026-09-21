package handlers

import (
	"fmt"
	"net/http"
	"SimilePro-go/internal/models"
	"time"
	"math/rand"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FiscalHandler struct {
	DB *gorm.DB
}

func NewFiscalHandler(db *gorm.DB) *FiscalHandler {
	return &FiscalHandler{DB: db}
}

// Helper
func (h *FiscalHandler) getClinicaID(c *gin.Context) (uint, error) {
	val, exists := c.Get("clinicaID")
	if !exists || val == nil {
		return 0, gorm.ErrRecordNotFound
	}
	return val.(uint), nil
}

// GetNotasFiscais godoc
// @Summary      Get invoices (NFS-e)
// @Description  List all service invoices for the clinic
// @Tags         fiscal
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.NotaFiscalServico
// @Router       /fiscal/nfs [get]
func (h *FiscalHandler) GetNotasFiscais(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	var notas []models.NotaFiscalServico
	if err := h.DB.Where("clinica_id = ?", clinicaID).Order("created_at DESC").Find(&notas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	c.JSON(http.StatusOK, notas)
}

// EmitirNFReq payload
type EmitirNFReq struct {
	FaturaID *uint   `json:"faturaId"`
	Valor    float64 `json:"valor"`
	Servico  string  `json:"servico"`
}

// EmitirNotaFiscal godoc
// @Summary      Issue NFS-e
// @Description  Request the issuance of a service invoice
// @Tags         fiscal
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body EmitirNFReq true "Invoice Request"
// @Success      201 {object} models.NotaFiscalServico
// @Router       /fiscal/nfs/emitir [post]
func (h *FiscalHandler) EmitirNotaFiscal(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	var req EmitirNFReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida"})
		return
	}

	if req.Valor <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valor inválido"})
		return
	}

	// Criação do registro
	aliquota := 3.5 // Exemplo de ISS
	valorIss := req.Valor * (aliquota / 100)

	nota := models.NotaFiscalServico{
		ClinicaID:     clinicaID,
		FaturaID:      req.FaturaID,
		Status:        models.StatusNFSEmitida, // Pulando status processando para o MVP
		Numero:        fmt.Sprintf("NFS%d", time.Now().UnixNano()/1000000), // Random number
		Rps:           fmt.Sprintf("RPS-%d", rand.Intn(9000)+1000),
		Serie:         "1",
		IssRetido:     false,
		ValorTotal:    req.Valor,
		ValorBaseCalc: req.Valor,
		Aliquota:      aliquota,
		ValorIss:      valorIss,
		CodigoServico: "04.01",
		ItemLc116:     "Medicina e biomedicina",
		Descricao:     req.Servico,
		Protocolo:     fmt.Sprintf("PROT-%d", time.Now().Unix()),
	}
	
	now := time.Now()
	nota.AutorizacaoAt = &now

	if err := h.DB.Create(&nota).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao emitir nota fiscal"})
		return
	}

	c.JSON(http.StatusCreated, nota)
}
