package main
import (
"fmt"
"SimilePro-go/internal/config"
"SimilePro-go/internal/database"
)
func main() {
cfg, _ := config.LoadConfig()
db, _ := database.Connect(cfg)
var count int64
db.Table("produtos").Count(&count)
fmt.Printf("COUNT=%d\n", count)
var ids []uint
db.Table("produtos").Distinct("clinica_id").Pluck("clinica_id", &ids)
fmt.Printf("IDS=%v\n", ids)
}
