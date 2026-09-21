package handlers

import (
	"fmt"
	"net/http"
	"SimilePro-go/internal/models"
	"SimilePro-go/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketHandler struct {
	DB *gorm.DB
}

func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{DB: db}
}

type CreateTicketRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required,max=2000"`
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	// Calculate sequence number (total tickets existing)
	var count int64
	if err := h.DB.Model(&models.Ticket{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar protocolo"})
		return
	}

	// Generate Protocol: 1 + 0000X + DDMMAAAA + HHMM
	// Format string for Go time: 02 (DD) 01 (MM) 2006 (AAAA) 15 (HH) 04 (MM)
	now := time.Now()
	timeStr := now.Format("020120061504")
	protocol := fmt.Sprintf("1%05d%s", count, timeStr)

	ticket := models.Ticket{
		Protocol: protocol,
		Email:    req.Email,
		Subject:  req.Subject,
		Message:  req.Message,
	}

	if err := h.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar ticket"})
		return
	}

	// Dispara o envio dos e-mails reais de forma assíncrona
	go func(email, proto, subj, msg string) {
		// 1. E-mail para o cliente (Remetente)
		clienteSubj := fmt.Sprintf("Cópia do Ticket [%s] - %s", proto, subj)
		clienteBody := fmt.Sprintf("Olá!\n\nRecebemos sua solicitação de suporte.\n\nGuarde o protocolo do seu ticket: %s\n\nSua mensagem original:\n\"%s\"\n\nNossa equipe técnica responderá em breve!", proto, msg)
		services.SendEmail([]string{email}, clienteSubj, clienteBody)

		// 2. E-mail para a equipe de suporte (Interno)
		suporteSubj := fmt.Sprintf("NOVO TICKET [%s] - %s", proto, subj)
		suporteBody := fmt.Sprintf("🚨 NOVO TICKET ABERTO\n\nProtocolo: %s\nDe: %s\nAssunto: %s\n\nMensagem do Usuário:\n\"%s\"\n", proto, email, subj, msg)
		services.SendEmail([]string{"similepro2@gmail.com"}, suporteSubj, suporteBody)
	}(req.Email, protocol, req.Subject, req.Message)

	c.JSON(http.StatusCreated, ticket)
}
