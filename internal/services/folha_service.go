package services

import (
	"errors"
	"odonto-flow-go/internal/models"
	"time"

	"gorm.io/gorm"
)

type FolhaService struct {
	db *gorm.DB
}

func NewFolhaService(db *gorm.DB) *FolhaService {
	return &FolhaService{db: db}
}

// CreateRubrica cria uma nova rubrica configurável para a clínica
func (s *FolhaService) CreateRubrica(clinicaID uint, rubrica *models.RubricaFolha) error {
	rubrica.ClinicaID = clinicaID
	return s.db.Create(rubrica).Error
}

// GetRubricas retorna as rubricas ativas da clínica
func (s *FolhaService) GetRubricas(clinicaID uint) ([]models.RubricaFolha, error) {
	var rubricas []models.RubricaFolha
	err := s.db.Where("clinica_id = ? AND ativa = ?", clinicaID, true).Find(&rubricas).Error
	return rubricas, err
}

// ProcessarFechamentoMes realiza o cálculo consolidado da folha
func (s *FolhaService) ProcessarFechamentoMes(clinicaID uint, mes, ano int) (*models.CompetenciaFolha, error) {
	var competencia models.CompetenciaFolha
	err := s.db.Where("clinica_id = ? AND mes = ? AND ano = ?", clinicaID, mes, ano).First(&competencia).Error

	if err == nil && competencia.Status == models.StatusFolhaPaga {
		return nil, errors.New("a folha desta competência já está encerrada e paga")
	}

	// Inicia transação para garantir atomicidade
	return &competencia, s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Limpar processamentos anteriores se existirem (Soft delete ou Hard delete dependendo da política)
		if competencia.ID != 0 {
			tx.Where("competencia_folha_id = ?", competencia.ID).Delete(&models.Holerite{})
		} else {
			competencia = models.CompetenciaFolha{
				ClinicaID: clinicaID,
				Mes:       mes,
				Ano:       ano,
				Status:    models.StatusFolhaAberta,
			}
			if err := tx.Create(&competencia).Error; err != nil {
				return err
			}
		}

		// 2. Buscar funcionários ativos
		var funcionarios []models.Funcionario
		tx.Where("clinica_id = ? AND status = ?", clinicaID, models.StatusFuncionarioAtivo).Find(&funcionarios)

		var totalBase, totalLiq, totalInss float64

		// 3. Processar cada funcionário
		for _, f := range funcionarios {
			holerite, err := s.calcularHoleriteFuncionario(tx, &competencia, &f)
			if err != nil {
				return err
			}
			
			totalBase += holerite.SalarioBase
			totalLiq += holerite.SalarioLiquido
			totalInss += (holerite.TotalProventos - holerite.SalarioLiquido) // Simplificado
		}

		// 4. Buscar dentistas para calcular comissões
		var dentistas []models.Dentista
		tx.Where("clinica_id = ?", clinicaID).Find(&dentistas)
		// ... lógica similar para dentistas

		// Atualizar totais da competência
		competencia.TotalBase = totalBase
		competencia.TotalLiq = totalLiq
		competencia.TotalInss = totalInss

		return tx.Save(&competencia).Error
	})
}

func (s *FolhaService) calcularHoleriteFuncionario(tx *gorm.DB, comp *models.CompetenciaFolha, f *models.Funcionario) (*models.Holerite, error) {
	salarioBase := f.Salario
	if salarioBase == 0 {
		salarioBase = 1412.00 // Mínimo
	}

	// 1. Garantir que existe a rubrica de Salário Base
	var rubricaBase models.RubricaFolha
	if err := tx.Where("clinica_id = ? AND descricao = ?", comp.ClinicaID, "Salário Base").First(&rubricaBase).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rubricaBase = models.RubricaFolha{
				ClinicaID:  comp.ClinicaID,
				Descricao:  "Salário Base",
				Tipo:       models.TipoEventoProvento,
				IncideINSS: true,
				IncideFGTS: true,
				IncideIRRF: true,
				Ativa:      true,
			}
			if err := tx.Create(&rubricaBase).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	holerite := models.Holerite{
		CompetenciaFolhaID: comp.ID,
		FuncionarioID:      f.ID,
		DiasTrabalhados:    30,
		SalarioBase:        salarioBase,
	}

	if err := tx.Create(&holerite).Error; err != nil {
		return nil, err
	}

	// 2. Lançar Salário Base como primeiro evento
	eventoBase := models.EventoHolerite{
		HoleriteID: holerite.ID,
		RubricaID:  rubricaBase.ID,
		Descricao:  rubricaBase.Descricao,
		Tipo:       rubricaBase.Tipo,
		Referencia: "30 Dias",
		Valor:      salarioBase,
	}
	if err := tx.Create(&eventoBase).Error; err != nil {
		return nil, err
	}

	// 3. Comissões (se houver)
	comissoes, _ := s.CalcularComissoesDentista(tx, f.UsuarioID, comp.Mes, comp.Ano) 
	if comissoes > 0 {
		// Tentar achar rubrica de comissão
		var rubricaComissao models.RubricaFolha
		tx.Where("clinica_id = ? AND descricao LIKE ?", comp.ClinicaID, "%Comissão%").First(&rubricaComissao)
		if rubricaComissao.ID == 0 {
			rubricaComissao = models.RubricaFolha{
				ClinicaID: comp.ClinicaID,
				Descricao: "Comissões",
				Tipo:      models.TipoEventoProvento,
				Ativa:     true,
			}
			tx.Create(&rubricaComissao)
		}

		tx.Create(&models.EventoHolerite{
			HoleriteID: holerite.ID,
			RubricaID:  rubricaComissao.ID,
			Descricao:  "Comissões sobre Procedimentos",
			Tipo:       models.TipoEventoProvento,
			Valor:      comissoes,
		})
	}

	// 4. Atualizar totais do holerite
	holerite.TotalProventos = salarioBase + comissoes
	holerite.SalarioLiquido = holerite.TotalProventos // Deduções seriam feitas aqui
	
	if err := tx.Save(&holerite).Error; err != nil {
		return nil, err
	}

	return &holerite, nil
}

func (s *FolhaService) CalcularComissoesDentista(tx *gorm.DB, dentistaID uint, mes, ano int) (float64, error) {
	// 1. Buscar dentista para pegar a porcentagem
	var dentista models.Dentista
	if err := tx.First(&dentista, "usuario_id = ?", dentistaID).Error; err != nil {
		// Se não for dentista, retorna 0
		return 0, nil
	}

	// 2. Buscar agendamentos atendidos no mês
	startDate := time.Date(ano, time.Month(mes), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var agendamentos []models.Agendamento
	tx.Where("dentista_id = ? AND status = ? AND data_hora_inicio >= ? AND data_hora_inicio < ?", 
		dentista.ID, models.StatusAgendamentoAtendido, startDate, endDate).Find(&agendamentos)

	// Nota: Em um sistema real, buscaríamos os itens de plano de tratamento vinculados
	// Aqui vamos simular um valor fixo por atendimento para demonstração da integração
	valorTotalProcedimentos := float64(len(agendamentos)) * 150.00 // Mock value

	comissao := (valorTotalProcedimentos * dentista.PorcentagemRepasse) / 100
	return comissao, nil
}
