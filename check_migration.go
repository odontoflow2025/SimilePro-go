package main
import (
"fmt"
"odonto-flow-go/internal/config"
"odonto-flow-go/internal/database"
"odonto-flow-go/internal/models"
)
func main() {
cfg, _ := config.LoadConfig()
db, _ := database.Connect(cfg)
err := db.AutoMigrate(&models.Produto{}, &models.NotaFiscalEntrada{}, &models.ItemNF{}, &models.MovimentacaoEstoque{})
if err != nil {
tln("MIGRATION_ERROR:", err)
} else {
tln("MIGRATION_SUCCESS")
}
}
