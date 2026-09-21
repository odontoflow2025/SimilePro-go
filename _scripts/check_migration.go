package main
import (
"fmt"
"SimilePro-go/internal/config"
"SimilePro-go/internal/database"
"SimilePro-go/internal/models"
)
func main() {
cfg, _ := config.LoadConfig()
db, _ := database.Connect(cfg)
err := db.AutoMigrate(&models.Produto{}, &models.NotaFiscalEntrada{}, &models.ItemNF{}, &models.MovimentacaoEstoque{})
if err != nil {
		fmt.Println("MIGRATION_ERROR:", err)
	} else {
		fmt.Println("MIGRATION_SUCCESS")
	}
}
