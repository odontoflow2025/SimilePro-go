package services

import (
	"context"
	"fmt"
)

// PaymentService defines the contract for interacting with any Payment Gateway
type PaymentService interface {
	// GenerateCharge creates a new charge intention (e.g. PIX or Credit Card)
	GenerateCharge(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
}

// PaymentRequest holds common data for gateway processing
type PaymentRequest struct {
	FaturaID   uint
	ClinicaID  uint
	Amount     float64
	Method     string // "PIX", "CREDIT_CARD", "BOLETO"
}

// PaymentResponse holds the payload to return to the Frontend
type PaymentResponse struct {
	TransactionID string
	ExternalID    string
	QrCodeData    string // For PIX
	Status        string
}

// externalIDBuilder builds the standard representation string mapping our local items to a Gateway
func BuildExternalID(clinicaID uint, faturaID uint) string {
	return fmt.Sprintf("CLN-%d-FAT-%d", clinicaID, faturaID)
}
