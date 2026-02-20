package services

import (
	"encoding/json"
	"log"
	"odonto-flow-go/internal/models"

	"gorm.io/gorm"
)

type AuditService struct {
    DB *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
    return &AuditService{DB: db}
}

func (s *AuditService) Log(usuarioID uint, clinicaID uint, acao models.AuditAction, recurso string, recursoID string, dadosAnteriores interface{}, dadosNovos interface{}, ipAddress string, userAgent string) {
    var dadosAnterioresStr, dadosNovosStr string

    if dadosAnteriores != nil {
        if b, err := json.Marshal(dadosAnteriores); err == nil {
            dadosAnterioresStr = string(b)
        }
    }

    if dadosNovos != nil {
        if b, err := json.Marshal(dadosNovos); err == nil {
            dadosNovosStr = string(b)
        }
    }

    logEntry := models.AuditLog{
        UsuarioID:       usuarioID,
        ClinicaID:       clinicaID,
        Acao:            acao,
        Recurso:         recurso,
        RecursoID:       recursoID,
        DadosAnteriores: dadosAnterioresStr,
        DadosNovos:      dadosNovosStr,
        IPAddress:       ipAddress,
        UserAgent:       userAgent,
    }

    // Run in background to not block request
    go func() {
        if err := s.DB.Create(&logEntry).Error; err != nil {
            log.Printf("Failed to create audit log: %v", err)
        }
    }()
}
