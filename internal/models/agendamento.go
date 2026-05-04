package models

import (
	"time"

	"gorm.io/gorm"
)

type StatusAgendamento string

const (
    StatusAgendamentoAgendado      StatusAgendamento = "AGENDADO"
    StatusAgendamentoConfirmado    StatusAgendamento = "CONFIRMADO"
    StatusAgendamentoCancelado     StatusAgendamento = "CANCELADO"
    StatusAgendamentoAtendido      StatusAgendamento = "ATENDIDO"
    StatusAgendamentoNaoCompareceu StatusAgendamento = "NAO_COMPARECEU"
)

type Agendamento struct {
    ID             uint              `gorm:"primaryKey" json:"id"`
    PacienteID     uint              `gorm:"not null" json:"pacienteId"`
    Paciente       Paciente          `gorm:"foreignKey:PacienteID" json:"paciente,omitempty"`
    DentistaID     uint              `gorm:"not null" json:"dentistaId"`
    Dentista       Dentista          `gorm:"foreignKey:DentistaID" json:"dentista,omitempty"`
    ClinicaID      uint              `gorm:"not null" json:"clinicaId"`
    Clinica        Clinica           `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
    DataHoraInicio time.Time         `json:"dataHoraInicio"`
    DataHoraFim    time.Time         `json:"dataHoraFim"`
    Motivo         string            `json:"motivoConsulta"`
    Status         StatusAgendamento `json:"status" gorm:"default:AGENDADO"`
    UsuarioCriacaoID uint            `json:"usuarioCriacaoId"`
    CreatedAt      time.Time         `json:"createdAt"`
    UpdatedAt      time.Time         `json:"updatedAt"`
    DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`
}
