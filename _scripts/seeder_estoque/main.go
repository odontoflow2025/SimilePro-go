package main

import (
	"fmt"
	"log"
	"math/rand"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/database"
	"odonto-flow-go/internal/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	var clinicas []models.Clinica
	if err := db.Find(&clinicas).Error; err != nil {
		log.Fatalf("Failed to fetch clinics: %v", err)
	}

    if len(clinicas) == 0 {
        log.Fatalf("No clinics found in DB!")
    }

	produtosTemplate := []models.Produto{
		{Nome: "Resina Composta Z350 XT A2", SKU: "RES-Z350-A2", UnidadeMedida: "UN", PrecoCusto: 85.50, EstoqueMinimo: 10},
		{Nome: "Anestésico Lidocaína 2% com Epinefrina", SKU: "ANE-LIDO-EPI", UnidadeMedida: "CX", PrecoCusto: 65.00, EstoqueMinimo: 5},
		{Nome: "Luva de Procedimento Látex P (100 un)", SKU: "LUV-LAT-P", UnidadeMedida: "CX", PrecoCusto: 32.90, EstoqueMinimo: 20},
		{Nome: "Luva de Procedimento Látex M (100 un)", SKU: "LUV-LAT-M", UnidadeMedida: "CX", PrecoCusto: 32.90, EstoqueMinimo: 20},
		{Nome: "Luva de Procedimento Látex G (100 un)", SKU: "LUV-LAT-G", UnidadeMedida: "CX", PrecoCusto: 32.90, EstoqueMinimo: 15},
		{Nome: "Máscara Tripla Descartável (50 un)", SKU: "MAS-TRI-50", UnidadeMedida: "CX", PrecoCusto: 18.50, EstoqueMinimo: 15},
		{Nome: "Sugador Descartável (40 un)", SKU: "SUG-DES-40", UnidadeMedida: "PCT", PrecoCusto: 12.00, EstoqueMinimo: 30},
		{Nome: "Rolo de Algodão Odontológico (500 un)", SKU: "ALG-ROL-500", UnidadeMedida: "PCT", PrecoCusto: 22.00, EstoqueMinimo: 10},
		{Nome: "Agulha Gengival Longa (100 un)", SKU: "AGU-GEN-LON", UnidadeMedida: "CX", PrecoCusto: 45.00, EstoqueMinimo: 5},
		{Nome: "Agulha Gengival Curta (100 un)", SKU: "AGU-GEN-CUR", UnidadeMedida: "CX", PrecoCusto: 45.00, EstoqueMinimo: 5},
		{Nome: "Gaze Odontológica Não Estéril (500 un)", SKU: "GAZ-NAO-EST", UnidadeMedida: "PCT", PrecoCusto: 15.50, EstoqueMinimo: 20},
		{Nome: "Adesivo Ambar Universal APS", SKU: "ADE-AMB-UNI", UnidadeMedida: "UN", PrecoCusto: 110.00, EstoqueMinimo: 3},
		{Nome: "Ácido Fosfórico 37% (3 Seringas)", SKU: "ACI-FOS-37", UnidadeMedida: "KIT", PrecoCusto: 28.00, EstoqueMinimo: 8},
		{Nome: "Cimento Ionômero de Vidro Ketac Molar", SKU: "CIM-ION-KET", UnidadeMedida: "KIT", PrecoCusto: 195.00, EstoqueMinimo: 2},
		{Nome: "Fio de Sutura Nylon 4-0 (24 un)", SKU: "SUT-NYL-40", UnidadeMedida: "CX", PrecoCusto: 75.00, EstoqueMinimo: 4},
	}

    fornecedores := []string{"Dental Cremer", "Surya Dental", "Dental Speed", "Ident"}
    baseDate := time.Now().AddDate(0, -1, 0) // Start inserting from a month ago

    for _, clinica := range clinicas {
        createdProdutos := make([]models.Produto, 0)
        
        for _, pTemplate := range produtosTemplate {
            p := pTemplate
            p.ClinicaID = clinica.ID
            var existing models.Produto
            if err := db.Where("clinica_id = ? AND sku = ?", clinica.ID, p.SKU).First(&existing).Error; err != nil {
                if err := db.Create(&p).Error; err != nil {
                    continue
                }
                createdProdutos = append(createdProdutos, p)
            } else {
                createdProdutos = append(createdProdutos, existing)
            }
        }
        
        log.Printf("Clinica %d has %d produtos", clinica.ID, len(createdProdutos))

        if len(createdProdutos) == 0 {
            continue
        }

        // Add 15 NFs for this clinic
        for i := 0; i < 15; i++ {
            prod := createdProdutos[i%len(createdProdutos)]
            qtd := float64(rand.Intn(50) + 10)
            precoTotal := qtd * prod.PrecoCusto

            nfNum := fmt.Sprintf("000%03d-%d", rand.Intn(9999), clinica.ID)
            forn := fornecedores[rand.Intn(len(fornecedores))]
            nfDate := baseDate.AddDate(0, 0, rand.Intn(30))

            // check if NF exists
            var existingNF models.NotaFiscalEntrada
            if err := db.Where("clinica_id = ? AND numero = ?", clinica.ID, nfNum).First(&existingNF).Error; err == nil {
                continue // already exists
            }

            err = db.Transaction(func(tx *gorm.DB) error {
                trans := models.Transacao{
                    ClinicaID:      clinica.ID,
                    Descricao:      "NF Entrada: " + nfNum + " - " + forn,
                    Valor:          precoTotal,
                    Tipo:           models.TipoTransacaoDespesa,
                    Status:         models.StatusTransacaoPago,
                    DataVencimento: nfDate.AddDate(0,0,30),
                    DataPagamento:  &nfDate,
                    Categoria:      "Estoque",
                }
                if err := tx.Create(&trans).Error; err != nil { return err }

                nf := models.NotaFiscalEntrada{
                    ClinicaID:   clinica.ID,
                    Numero:      nfNum,
                    Fornecedor:  forn,
                    DataEmissao: nfDate,
                    ValorTotal:  precoTotal,
                    TransacaoID: &trans.ID,
                }
                if err := tx.Create(&nf).Error; err != nil { return err }

                itemNF := models.ItemNF{
                    NotaFiscalID:  nf.ID,
                    ProdutoID:     prod.ID,
                    Quantidade:    qtd,
                    PrecoUnitario: prod.PrecoCusto,
                    Subtotal:      precoTotal,
                }
                if err := tx.Create(&itemNF).Error; err != nil { return err }

                mov := models.MovimentacaoEstoque{
                    ClinicaID:    clinica.ID,
                    ProdutoID:    prod.ID,
                    Tipo:         models.MovimentacaoEntrada,
                    Quantidade:   qtd,
                    NotaFiscalID: &nf.ID,
                    Observacao:   "Entrada via NF " + nfNum,
                    CreatedAt:    nfDate,
                }
                if err := tx.Create(&mov).Error; err != nil { return err }

                if err := tx.Model(&models.Produto{}).Where("id = ? AND clinica_id = ?", prod.ID, clinica.ID).
                    UpdateColumn("estoque_atual", gorm.Expr("estoque_atual + ?", strconv.FormatFloat(qtd, 'f', -1, 64))).Error; err != nil {
                    return err
                }
                return nil
            })
        }
    }

    log.Println("Seeding Estoque completed for all clinics successfully!")
}
