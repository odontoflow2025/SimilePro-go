package main

import (
	"fmt"
	"log"
	"math/rand"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Falha ao carregar config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Falha ao conectar no banco: %v", err)
	}

	log.Println(">>> Iniciando Super Seeder de Estresse RH/Financeiro...")

	// 1. Identificar Clínica Ativa (ou criar se não existir)
	var clinica models.Clinica
	db.First(&clinica)
	if clinica.ID == 0 {
		log.Fatal("Nenhuma clínica encontrada. Por favor, cadastre uma clínica primeiro ou rode o seeder inicial.")
	}
	log.Printf("Operando na Clínica: %s (ID: %d)\n", clinica.NomeFantasia, clinica.ID)

	// 2. Criar Massa de Colaboradores (15 novos)
	cargos := []string{"Recepcionista", "Cirurgião-Dentista", "ASB", "TSB", "Financeiro", "Serviços Gerais"}
	var funcionarios []models.Funcionario

	for i := 1; i <= 15; i++ {
		email := fmt.Sprintf("colaborador%d@odontoflow.com", i + 100) // Evitar conflitos
		cpf := fmt.Sprintf("%03d%03d%03d%02d", rand.Intn(999), rand.Intn(999), rand.Intn(999), rand.Intn(99))
		
		user := models.User{
			Nome:        fmt.Sprintf("Colaborador Teste %d", i),
			Email:       email,
			CPF:         cpf,
			TipoUsuario: "FUNCIONARIO",
			SenhaHash:   hashPassword("mudar123"),
			ClinicaID:   clinica.ID,
		}
		db.Where("email = ?", email).FirstOrCreate(&user)

		dataAdm := time.Now().AddDate(-1, -rand.Intn(6), -rand.Intn(28))
		funcio := models.Funcionario{
			UsuarioID:    user.ID,
			ClinicaID:    clinica.ID,
			Cargo:        cargos[rand.Intn(len(cargos))],
			DataAdmissao: dataAdm,
			Salario:      2000.00 + (float64(i) * 150.0),
			PIS:          fmt.Sprintf("123.%05d.%03d-%d", i, rand.Intn(999), rand.Intn(9)),
			CTPS:         fmt.Sprintf("%07d/%04d", rand.Intn(9999999), rand.Intn(2025)),
			RG:           fmt.Sprintf("%d.%d.%d-%d", rand.Intn(99), rand.Intn(999), rand.Intn(999), rand.Intn(9)),
			Endereco:     "Rua dos Testes, 123 - Bairro Stress - Cidade/UF",
			DadosBanco:   "Banco Digital - Ag: 0001 - CC: 12345-6",
			CargaHoraria: 220,
			Status:       models.StatusFuncionarioAtivo,
		}
		db.Where("usuario_id = ?", user.ID).FirstOrCreate(&funcio)
		funcionarios = append(funcionarios, funcio)
	}
	log.Println("[OK] 15 Colaboradores processados.")

	// 3. Aplicar 8 Demissões (Rescisões)
	log.Println(">>> Processando 8 Desligamentos...")
	motivos := []string{"Pedido de Demissão", "Corte de Gastos", "Performance", "Término de Contrato"}
	for i := 0; i < 8; i++ {
		f := &funcionarios[i]
		dataDem := time.Now().AddDate(0, 0, -i)
		db.Model(f).Updates(map[string]interface{}{
			"status":          models.StatusFuncionarioDemitido,
			"data_demissao":   &dataDem,
			"motivo_demissao": motivos[rand.Intn(len(motivos))],
		})
	}
	log.Println("[OK] 8 Colaboradores demitidos com sucesso.")

	// 4. Notas Fiscais (Entrada e Saída)
	log.Println(">>> Gerando Movimentação Fiscal (Notas)...")
	for i := 0; i < 10; i++ {
		// Entrada (Compra)
		nfEntrada := models.NotaFiscalEntrada{
			ClinicaID:   clinica.ID,
			Numero:      fmt.Sprintf("NF-E-%d", 1000+i),
			Fornecedor:  "Fornecedor de Insumos Odonto S.A",
			DataEmissao: time.Now().AddDate(0, 0, -i),
			ValorTotal:  float64(500 + (i * 120)),
		}
		db.Create(&nfEntrada)

		// Saída (Serviço)
		nfSaida := models.NotaFiscalServico{
			ClinicaID:  clinica.ID,
			Numero:     fmt.Sprintf("NFS-%d", 5000+i),
			Descricao:  "Prestação de Serviços Odontológicos",
			ValorTotal: float64(200 + (i * 50)),
			Status:     models.StatusNFSEmitida,
		}
		db.Create(&nfSaida)
	}
	log.Println("[OK] 20 Notas Fiscais geradas.")

	// 5. Pagamentos Tributários (Guias de Impostos)
	log.Println(">>> Registrando Guias de Impostos...")
	impostos := []string{"INSS Patronal", "FGTS Mensal", "ISSQN", "IRRF Folha"}
	for _, imp := range impostos {
		guia := models.ImpostoGuia{
			ClinicaID:      clinica.ID,
			Descricao:      imp + " - Competência 04/2026",
			ValorTotal:     float64(300 + rand.Intn(1000)),
			DataVencimento: time.Now().AddDate(0, 0, 5),
			Status:         models.StatusFaturaPendente,
		}
		db.Create(&guia)
	}
	log.Println("[OK] Guias tributárias registradas.")

	// 6. Holerites (Folha de Pagamento)
	log.Println(">>> Simulando Holerites...")
	var comp models.CompetenciaFolha
	db.Where("clinica_id = ? AND mes = 5 AND ano = 2026", clinica.ID).First(&comp)
	if comp.ID == 0 {
		comp = models.CompetenciaFolha{
			ClinicaID: clinica.ID,
			Mes:       5,
			Ano:       2026,
			Status:    models.StatusFolhaAberta,
		}
		db.Create(&comp)
	}

	for _, f := range funcionarios {
		hol := models.Holerite{
			CompetenciaFolhaID: comp.ID,
			FuncionarioID:      f.ID,
			DiasTrabalhados:    30,
			SalarioBase:        f.Salario,
			TotalProventos:     f.Salario,
			SalarioLiquido:     f.Salario * 0.9, // Simula descontos
		}
		db.Create(&hol)
	}
	log.Println("[OK] Massa de Holerites gerada.")

	log.Println("\n>>> SUPER SEEDER CONCLUÍDO COM SUCESSO! <<<")
	log.Println("Acesse o painel de RH para ver as 8 demissões e a nova folha.")
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}
