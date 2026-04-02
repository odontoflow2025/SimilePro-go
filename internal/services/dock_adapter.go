package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// OdontoflowGatewayAdapter is the HTTP client pointing to the external microservice gateway-odontoflow
type OdontoflowGatewayAdapter struct {
	BaseURL string
	Client  *http.Client
}

// NewDockMockAdapter initializes a connection to the microservice
func NewDockMockAdapter() PaymentService {
	log.Println("[INFO] Initializing OdontoflowGatewayAdapter to link with gateway-odontoflow...")
	return &OdontoflowGatewayAdapter{
		BaseURL: "http://localhost:8081", // Assuming gateway runs on 8081 to avoid 8080 clash
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GenerateCharge calls the gateway-odontoflow POST /api/v1/payments/init
func (d *OdontoflowGatewayAdapter) GenerateCharge(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	log.Printf("[INFO] GatewayAdapter: Sending charge to microservice for Amount: %.2f\n", req.Amount)

	externalID := BuildExternalID(req.ClinicaID, req.FaturaID)

	payload := map[string]interface{}{
		"dentist_id":     req.ClinicaID, // Mapping Clinica to Dentist conceptually for the microservice
		"amount":         req.Amount,
		"payment_method": req.Method,
		"external_id":    externalID,
	}

	bodyBytes, _ := json.Marshal(payload)
	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/payments/init", d.BaseURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build gateway request: %v", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to call gateway microservice: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gateway returned error status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gateway returned invalid JSON: %v", err)
	}

	log.Printf("[INFO] GatewayAdapter: Successfully connected to microservice. ExternalID: %s\n", externalID)

	// Return response parsing from the gateway format
	return &PaymentResponse{
		TransactionID: fmt.Sprintf("%v", result["transaction_id"]),
		ExternalID:    externalID,
		QrCodeData:    fmt.Sprintf("%v", result["qr_code"]),
		Status:        "pending",
	}, nil
}
