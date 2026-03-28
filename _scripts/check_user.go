package main
import (
"fmt"
"odonto-flow-go/internal/config"
"odonto-flow-go/internal/database"
)
func main() {
cfg, _ := config.LoadConfig()
db, _ := database.Connect(cfg)
var users []map[string]interface{}
db.Table("users").Limit(5).Find(&users)
for _, u := range users {
tf("USER: ID=%v, Nome=%v, ClinicaID=%v\n", u["id"], u["nome"], u["clinica_id"])
}
}
