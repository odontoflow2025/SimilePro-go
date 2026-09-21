package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"SimilePro-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DockWebhookPayload represents the mock payload we expect backwards from Dock
type DockWebhookPayload struct {
	EventID       string  `json:"event_id"`
	TransactionID string  `json:"transaction_id"`
	ExternalID    string  `json:"external_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
}

// HandleDockWebhook processes incoming updates from the Dock gateway
func HandleDockWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload DockWebhookPayload

		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Printf("[WEBHOOK-ERR] Invalid Dock payload: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
			return
		}

		log.Printf("[WEBHOOK-INFO] Received Dock Update. ExtID: %s, Status: %s\n", payload.ExternalID, payload.Status)

		// 1. Authenticate Request
		// In a real scenario, we'd check `X-Dock-Signature` HMAC. Since this is Mock phase:
		authHeader := c.GetHeader("X-Dock-Signature")
		if authHeader == "" {
			log.Println("[WEBHOOK-WARN] Missing X-Dock-Signature. Accepting anyway because we are in Mock mode.")
		}

		// 2. Parse Context (ExternalID format: CLN-{clinicaID}-FAT-{faturaID})
		clinicaID, faturaID, err := parseExternalID(payload.ExternalID)
		if err != nil {
			log.Printf("[WEBHOOK-ERR] Cannot parse ExternalID (%s): %v\n", payload.ExternalID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown external_id pattern"})
			return
		}

		// 3. Process only "paid" events
		if payload.Status == "paid" {
			err = processPayment(db, clinicaID, faturaID, payload.Amount, payload.TransactionID)
			if err != nil {
				log.Printf("[WEBHOOK-ERR] Failed to process DB changes: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal processor error"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
	}
}

// processPayment wraps the status changing and ledger creation in a Transaction
func processPayment(db *gorm.DB, clinicaID uint, faturaID uint, amount float64, gatewayTxID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// A. Fetch invoice
		var fatura models.Fatura
		if err := tx.Where("id = ? AND clinica_id = ?", faturaID, clinicaID).First(&fatura).Error; err != nil {
			return fmt.Errorf("Fatura %d not found for Clinica %d", faturaID, clinicaID)
		}

		// Skip if already paid
		if fatura.Status == models.StatusFaturaPaga {
			return nil
		}

		// B. Update invoice status
		if err := tx.Model(&fatura).Updates(map[string]interface{}{
			"status": models.StatusFaturaPaga,
		}).Error; err != nil {
			return err
		}

		// C. Generate related general ledger Transaction
		now := time.Now()
		transacao := models.Transacao{
			ClinicaID:      fatura.ClinicaID,
			PacienteID:     &fatura.PacienteID,
			Descricao:      fmt.Sprintf("Pagamento via Dock Webhook (TxID: %s)", gatewayTxID),
			Valor:          amount,
			Tipo:           models.TipoTransacaoReceita,
			Status:         models.StatusTransacaoPago,
			DataVencimento: fatura.DataVencimento,
			DataPagamento:  &now,
			Categoria:      "Pagamentos Checkout",
			FormaPagamento: "Gateway (Dock)",
		}

		if err := tx.Create(&transacao).Error; err != nil {
			return err
		}

		log.Printf("[WEBHOOK-SUCCESS] Fatura %d marked as PAID. Internal Tx %d created.", fatura.ID, transacao.ID)
		return nil
	})
}

// parseExternalID parses generic mapping strings
func parseExternalID(extID string) (uint, uint, error) {
	re := regexp.MustCompile(`^CLN-(\d+)-FAT-(\d+)$`)
	matches := re.FindStringSubmatch(extID)

	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid format")
	}

	cID, _ := strconv.Atoi(matches[1])
	fID, _ := strconv.Atoi(matches[2])

	return uint(cID), uint(fID), nil
}
