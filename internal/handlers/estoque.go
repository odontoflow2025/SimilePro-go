package handlers

import (
	"fmt"
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EstoqueHandler struct {
	DB *gorm.DB
}

func NewEstoqueHandler(db *gorm.DB) *EstoqueHandler {
	return &EstoqueHandler{DB: db}
}

// GetProdutos retorna a lista de produtos da clínica
func (h *EstoqueHandler) GetProdutos(c *gin.Context) {
	fmt.Println("DEBUG: GetProdutos called")
	clinicaIDRaw, _ := c.Get("clinicaID")
	var clinicaID uint
	switch v := clinicaIDRaw.(type) {
	case uint:
		clinicaID = v
	case float64:
		clinicaID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID da clínica inválido"})
		return
	}
	produtos := []models.Produto{}
	if err := h.DB.Where("clinica_id = ?", clinicaID).Find(&produtos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar produtos"})
		return
	}
	c.JSON(http.StatusOK, produtos)
}

// CreateNotaFiscal lança uma NF, atualiza estoque e gera despesa financeira
func (h *EstoqueHandler) CreateNotaFiscal(c *gin.Context) {
	clinicaIDRaw, _ := c.Get("clinicaID")
	var clinicaID uint
	switch v := clinicaIDRaw.(type) {
	case uint:
		clinicaID = v
	case float64:
		clinicaID = uint(v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID da clínica inválido"})
		return
	}

	var input struct {
		Numero      string    `json:"numero" binding:"required"`
		Fornecedor  string    `json:"fornecedor" binding:"required"`
		DataEmissao time.Time `json:"dataEmissao"`
		ValorTotal  float64   `json:"valorTotal" binding:"required"`
		Itens       []struct {
			ProdutoID     uint    `json:"produtoId" binding:"required"`
			Quantidade    float64 `json:"quantidade" binding:"required"`
			PrecoUnitario float64 `json:"precoUnitario" binding:"required"`
		} `json:"itens" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Inicia transação no banco de dados
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Criar Transação Financeira (Despesa)
		transacao := models.Transacao{
			ClinicaID:      clinicaID,
			Descricao:      "NF Entrada: " + input.Numero + " - " + input.Fornecedor,
			Valor:          input.ValorTotal,
			Tipo:           models.TipoTransacaoDespesa,
			Status:         models.StatusTransacaoPendente,
			DataVencimento: time.Now().AddDate(0, 0, 30), // Vencimento padrão 30 dias
			Categoria:      "Estoque",
		}
		if err := tx.Create(&transacao).Error; err != nil {
			return err
		}

		// 2. Criar Nota Fiscal de Entrada
		nf := models.NotaFiscalEntrada{
			ClinicaID:   clinicaID,
			Numero:      input.Numero,
			Fornecedor:  input.Fornecedor,
			DataEmissao: input.DataEmissao,
			ValorTotal:  input.ValorTotal,
			TransacaoID: &transacao.ID,
		}
		if err := tx.Create(&nf).Error; err != nil {
			return err
		}

		// 3. Processar Itens, Movimentação e Estoque
		for _, itemInput := range input.Itens {
			// Criar Item da NF
			itemNF := models.ItemNF{
				NotaFiscalID:  nf.ID,
				ProdutoID:     itemInput.ProdutoID,
				Quantidade:    itemInput.Quantidade,
				PrecoUnitario: itemInput.PrecoUnitario,
				Subtotal:      itemInput.Quantidade * itemInput.PrecoUnitario,
			}
			if err := tx.Create(&itemNF).Error; err != nil {
				return err
			}

			// Criar Movimentação de Estoque
			mov := models.MovimentacaoEstoque{
				ClinicaID:    clinicaID,
				ProdutoID:    itemInput.ProdutoID,
				Tipo:         models.MovimentacaoEntrada,
				Quantidade:   itemInput.Quantidade,
				NotaFiscalID: &nf.ID,
				Observacao:   "Entrada via NF " + nf.Numero,
				CreatedAt:    time.Now(),
			}
			if err := tx.Create(&mov).Error; err != nil {
				return err
			}

			// Atualizar Saldo do Produto
			if err := tx.Model(&models.Produto{}).Where("id = ? AND clinica_id = ?", itemInput.ProdutoID, clinicaID).
				UpdateColumn("estoque_atual", gorm.Expr("estoque_atual + ?", itemInput.Quantidade)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar Nota Fiscal: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Nota Fiscal lançada com sucesso, estoque atualizado e despesa gerada."})
}
