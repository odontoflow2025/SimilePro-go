package handlers

import (
	"log"
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FolhaHandler struct {
	DB *gorm.DB
}

func NewFolhaHandler(db *gorm.DB) *FolhaHandler {
	return &FolhaHandler{DB: db}
}

// Helper para obter clinicaID do contexto
func (h *FolhaHandler) getClinicaID(c *gin.Context) (uint, error) {
	return getClinicaIDFromContext(c)
}

// Request payload
type ProcessarFolhaReq struct {
	Mes int `json:"mes"`
	Ano int `json:"ano"`
}

// ProcessarFolha processing logic
func (h *FolhaHandler) ProcessarFolha(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	var req ProcessarFolhaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida"})
		return
	}

	// 1. Verificar se a folha já existe e não está PAGA
	var competencia models.CompetenciaFolha
	err = h.DB.Where("clinica_id = ? AND mes = ? AND ano = ?", clinicaID, req.Mes, req.Ano).First(&competencia).Error

	if err == nil {
		if competencia.Status == models.StatusFolhaPaga {
			c.JSON(http.StatusConflict, gin.H{"error": "A folha desta competência já está encerrada/paga."})
			return
		}
		// Vamos excluir os holerites antigos para recalcular
		h.DB.Where("competencia_folha_id = ?", competencia.ID).Delete(&models.Holerite{})
		h.DB.Where("id = ?", competencia.ID).Delete(&models.CompetenciaFolha{})
	}

	// 2. Buscar todos os funcionários ativos
	var funcionarios []models.Funcionario
	h.DB.Where("clinica_id = ? AND status = ?", clinicaID, "ATIVO").Find(&funcionarios)

	if len(funcionarios) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nenhum funcionário ativo para processar a folha."})
		return
	}

	// 3. Criar nova competência base
	novaCompetencia := models.CompetenciaFolha{
		ClinicaID: clinicaID,
		Mes:       req.Mes,
		Ano:       req.Ano,
		Status:    models.StatusFolhaAberta,
	}

	// Inicia transação
	tx := h.DB.Begin()

	if err := tx.Create(&novaCompetencia).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar competência"})
		return
	}

	var totalBase, totalLiq, totalInss, totalIrrf float64

	// 4. Processar cada funcionário
	for _, funcio := range funcionarios {
		// Variáveis base
		salarioBase := funcio.Salario // Requer que SalarioBase seja adicionado real no model (está como float64)
		if salarioBase == 0 {
			salarioBase = 1412.00 // Salário mínimo fallback
		}

		inss := salarioBase * 0.09 // Simplificação fictícia da tabela progressiva
		irrf := 0.0
		if salarioBase > 2824.00 {
			irrf = (salarioBase * 0.15) - 380.00 // Simplificação fictícia
		}

		if irrf < 0 {
			irrf = 0
		}

		liquido := salarioBase - inss - irrf

		// Somatórios da competência
		totalBase += salarioBase
		totalInss += inss
		totalIrrf += irrf
		totalLiq += liquido

		// Cria as rubricas/eventos
		eventos := []models.EventoHolerite{
			{Descricao: "Salário Base", Tipo: models.TipoEventoProvento, Referencia: "30 Dias", Valor: salarioBase},
			{Descricao: "Desconto INSS", Tipo: models.TipoEventoDesconto, Referencia: "Tabela Padrão", Valor: inss},
		}

		if irrf > 0 {
			eventos = append(eventos, models.EventoHolerite{Descricao: "Desconto IRRF", Tipo: models.TipoEventoDesconto, Referencia: "Tabela Padrão", Valor: irrf})
		}

		// Save holerite
		holerite := models.Holerite{
			CompetenciaFolhaID: novaCompetencia.ID,
			FuncionarioID:      funcio.ID,
			DiasTrabalhados:    30,
			SalarioBase:        salarioBase,
			TotalProventos:     salarioBase,
			TotalDescontos:     inss + irrf,
			SalarioLiquido:     liquido,
			Eventos:            eventos,
		}

		if err := tx.Create(&holerite).Error; err != nil {
			log.Printf("Erro ao processar holerite: %v", err)
		}
	}

	// Atualiza totais
	tx.Model(&novaCompetencia).Updates(models.CompetenciaFolha{
		TotalBase: totalBase,
		TotalLiq:  totalLiq,
		TotalInss: totalInss,
		TotalIrrf: totalIrrf,
	})

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Folha processada com sucesso", "competencia": novaCompetencia})
}

// GetHolerites retorna a lista de holerites para a competência dada
func (h *FolhaHandler) GetHolerites(c *gin.Context) {
	clinicaID, err := h.getClinicaID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	mes := c.Query("mes")
	ano := c.Query("ano")

	var competencia models.CompetenciaFolha
	if err := h.DB.Where("clinica_id = ? AND mes = ? AND ano = ?", clinicaID, mes, ano).First(&competencia).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folha não encontrada para o período"})
		return
	}

	var holerites []models.Holerite
	h.DB.Preload("Funcionario").Preload("Eventos").Where("competencia_folha_id = ?", competencia.ID).Find(&holerites)

	c.JSON(http.StatusOK, gin.H{"competencia": competencia, "holerites": holerites})
}
