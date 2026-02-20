package models

import (
	"time"
)

type AuditAction string

const (
    AuditActionCreate AuditAction = "CREATE"
    AuditActionUpdate AuditAction = "UPDATE"
    AuditActionDelete AuditAction = "DELETE"
    AuditActionLogin  AuditAction = "LOGIN"
)

type AuditLog struct {
    ID             uint           `gorm:"primaryKey" json:"id"`
    UsuarioID      uint           `gorm:"not null" json:"usuarioId"`
    Usuario        User           `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
    ClinicaID      uint           `json:"clinicaId"`
    Clinica        *Clinica       `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    Acao           AuditAction    `gorm:"not null" json:"acao"`
    Recurso        string         `gorm:"not null" json:"recurso"` // Ex: "PACIENTE", "AGENDAMENTO"
    RecursoID      string         `json:"recursoId"`
    DadosAnteriores string         `gorm:"type:text" json:"dadosAnteriores"` // JSON string
    DadosNovos     string         `gorm:"type:text" json:"dadosNovos"`      // JSON string
    IPAddress      string         `json:"ipAddress"`
    UserAgent      string         `json:"userAgent"`
    CreatedAt      time.Time      `json:"createdAt"`
}
