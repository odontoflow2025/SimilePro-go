package handlers

import (
	"fmt"
	"net/http"
	"odonto-flow-go/internal/models"
	"odonto-flow-go/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RHFinanceiroHandler struct {
	DB *gorm.DB
}

func NewRHFinanceiroHandler(db *gorm.DB) *RHFinanceiroHandler {
	return &RHFinanceiroHandler{DB: db}
}

// --- SEÇÃO RH: Gestão de Profissionais & Valores ---

type UpsertFuncionarioReq struct {
	UsuarioID    uint      `json:"usuarioId" binding:"required"`
	Cargo        string    `json:"cargo" binding:"required"`
	DataAdmissao time.Time `json:"dataAdmissao"`
	SalarioBase  float64   `json:"salarioBase"`
	PIS          string    `json:"pis"`
	CTPS         string    `json:"ctps"`
	RG           string    `json:"rg"`
	DadosBanco   string    `json:"dadosBanco"`
}

func (h *RHFinanceiroHandler) SaveFuncionario(c *gin.Context) {
	clinicaID, _ := getClinicaIDFromContext(c)
	var req UpsertFuncionarioReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	var funcionario models.Funcionario
	res := h.DB.Where("usuario_id = ? AND clinica_id = ?", req.UsuarioID, clinicaID).First(&funcionario)

	funcionario.UsuarioID = req.UsuarioID
	funcionario.ClinicaID = clinicaID
	funcionario.Cargo = req.Cargo
	funcionario.DataAdmissao = req.DataAdmissao
	funcionario.Salario = req.SalarioBase
	funcionario.PIS = req.PIS
	funcionario.CTPS = req.CTPS
	funcionario.RG = req.RG
	funcionario.DadosBanco = req.DadosBanco

	// Nota: models.Funcionario possui hooks BeforeSave que chamam utils.EncryptAESGCM
	if res.Error != nil {
		if err := h.DB.Create(&funcionario).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar registro de RH"})
			return
		}
	} else {
		if err := h.DB.Save(&funcionario).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar registro de RH"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dados de RH salvos com sucesso (Blindagem Ativa)"})
}

// Edição Manual de Valores (Bônus/Descontos Avulsos)
type AjusteManualReq struct {
	HoleriteID uint    `json:"holeriteId" binding:"required"`
	RubricaID  uint    `json:"rubricaId" binding:"required"`
	Valor      float64 `json:"valor" binding:"required"`
	Referencia string  `json:"referencia"`
}

func (h *RHFinanceiroHandler) AplicarAjusteManual(c *gin.Context) {
	clinicaID, _ := getClinicaIDFromContext(c)
	var req AjusteManualReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Verificar se a holerite pertence à clínica
	var holerite models.Holerite
	if err := h.DB.Joins("Competencia").Where("holerites.id = ? AND \"Competencia\".clinica_id = ?", req.HoleriteID, clinicaID).First(&holerite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Holerite não encontrada ou acesso negado"})
		return
	}

	var rubrica models.RubricaFolha
	if err := h.DB.First(&rubrica, req.RubricaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rubrica não encontrada"})
		return
	}

	evento := models.EventoHolerite{
		HoleriteID: req.HoleriteID,
		RubricaID:  req.RubricaID,
		Descricao:  "(Ajuste Manual) " + rubrica.Descricao,
		Tipo:       rubrica.Tipo,
		Referencia: req.Referencia,
		Valor:      req.Valor,
	}

	if err := h.DB.Create(&evento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao aplicar ajuste"})
		return
	}

	// Recalcular totais da holerite
	h.recalcularHolerite(req.HoleriteID)

	c.JSON(http.StatusOK, gin.H{"message": "Ajuste aplicado e folha recalculada"})
}

// --- SEÇÃO FINANCEIRA: Tributos & Pagamentos ---

func (h *RHFinanceiroHandler) FecharFolhaEFiscal(c *gin.Context) {
	clinicaID, _ := getClinicaIDFromContext(c)
	mes, _ := strconv.Atoi(c.Query("mes"))
	ano, _ := strconv.Atoi(c.Query("ano"))

	if mes == 0 || ano == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mês e Ano são obrigatórios"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var competencia models.CompetenciaFolha
		if err := tx.Where("clinica_id = ? AND mes = ? AND ano = ?", clinicaID, mes, ano).First(&competencia).Error; err != nil {
			return fmt.Errorf("competência não encontrada")
		}

		if competencia.Status == models.StatusFolhaFechada {
			return fmt.Errorf("folha já está fechada")
		}

		// 1. Consolidar Provisão de Tributos (INSS, FGTS, IRRF)
		// Aqui utilizamos int64 (cents) para evitar drift de float64
		valorINSS := int64(competencia.TotalInss * 100)
		valorFGTS := int64(competencia.TotalFgts * 100)
		valorIRRF := int64(competencia.TotalIrrf * 100)

		tributos := []models.BalancoFinanceiroTenant{
			{ClinicaID: clinicaID, Mes: mes, Ano: ano, Tipo: "TRIBUTO_INSS", Descricao: "Encargos Previdenciários (Folha)", Valor: valorINSS, ReferenciaID: competencia.ID},
			{ClinicaID: clinicaID, Mes: mes, Ano: ano, Tipo: "TRIBUTO_FGTS", Descricao: "Fundo de Garantia (Folha)", Valor: valorFGTS, ReferenciaID: competencia.ID},
			{ClinicaID: clinicaID, Mes: mes, Ano: ano, Tipo: "TRIBUTO_IRRF", Descricao: "Imposto de Renda Retido (Folha)", Valor: valorIRRF, ReferenciaID: competencia.ID},
		}

		for _, t := range tributos {
			if t.Valor > 0 {
				if err := tx.Create(&t).Error; err != nil {
					return err
				}
			}
		}

		// 2. Marcar competência como fechada
		competencia.Status = models.StatusFolhaFechada
		return tx.Save(&competencia).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folha encerrada e tributos provisionados no Balanço Financeiro"})
}

// --- EXPORTAÇÃO ---

func (h *RHFinanceiroHandler) ExportarHolerite(c *gin.Context) {
	clinicaID, _ := getClinicaIDFromContext(c)
	holeriteID := c.Param("holerite_id")

	var holerite models.Holerite
	err := h.DB.Preload("Funcionario").Preload("Eventos").Preload("Competencia").
		Where("id = ?", holeriteID).First(&holerite).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Holerite inexistente"})
		return
	}

	// Segurança: Validar Tenant
	if holerite.Competencia.ClinicaID != clinicaID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Violação de perímetro: Holerite pertence a outra unidade"})
		return
	}

	// Descriptografar dados sensíveis do funcionário para o documento
	// Nota: Em AfterFind o modelo já tenta descriptografar, mas aqui garantimos a integridade
	h.decryptFuncionarioFields(&holerite.Funcionario)

	// Estrutura pronta para renderização PDF ou JSON Estruturado
	data := gin.H{
		"unidade":    holerite.Competencia.ClinicaID,
		"periodo":    fmt.Sprintf("%02d/%d", holerite.Competencia.Mes, holerite.Competencia.Ano),
		"funcionario": holerite.Funcionario.Usuario.Nome, // Assumindo Preload("Funcionario.Usuario")
		"cargo":      holerite.Funcionario.Cargo,
		"pis":        holerite.Funcionario.PIS,
		"ctps":       holerite.Funcionario.CTPS,
		"eventos":    holerite.Eventos,
		"total_prov": holerite.TotalProventos,
		"total_desc": holerite.TotalDescontos,
		"liquido":    holerite.SalarioLiquido,
	}

	c.JSON(http.StatusOK, data)
}

// Utils Internos

func (h *RHFinanceiroHandler) recalcularHolerite(id uint) {
	var holerite models.Holerite
	h.DB.Preload("Eventos").First(&holerite, id)

	var prov, desc float64
	for _, e := range holerite.Eventos {
		if e.Tipo == models.TipoEventoProvento {
			prov += e.Valor
		} else {
			desc += e.Valor
		}
	}
	holerite.TotalProventos = prov
	holerite.TotalDescontos = desc
	holerite.SalarioLiquido = prov - desc
	h.DB.Save(&holerite)
}

func (h *RHFinanceiroHandler) decryptFuncionarioFields(f *models.Funcionario) {
	// Se os hooks de models.Funcionario falharam ou queremos garantir agora:
	if f.PIS != "" && len(f.PIS) > 30 { // Simples check se está encodado (b64/hex)
		dec, _ := utils.DecryptAESGCM(f.PIS)
		if dec != "" { f.PIS = dec }
	}
	if f.CTPS != "" && len(f.CTPS) > 30 {
		dec, _ := utils.DecryptAESGCM(f.CTPS)
		if dec != "" { f.CTPS = dec }
	}
}
