package handlers

import (
	"net/http"
	"SimilePro-go/internal/models"
	"SimilePro-go/internal/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgendamentoHandler struct {
    DB *gorm.DB
}

func NewAgendamentoHandler(db *gorm.DB) *AgendamentoHandler {
    return &AgendamentoHandler{DB: db}
}

type CreateAgendamentoInput struct {
    PacienteID       uint      `json:"pacienteId" binding:"required"`
    DentistaID       uint      `json:"dentistaId" binding:"required"`
    DataHoraInicio   time.Time `json:"dataHoraInicio" binding:"required"`
    DataHoraFim      time.Time `json:"dataHoraFim" binding:"required"`
    Motivo           string    `json:"motivo"`
    UsuarioCriacaoID uint      `json:"usuarioCriacaoId"`
}

type UpdateAgendamentoInput struct {
    MotivoConsulta *string                   `json:"motivoConsulta"`
    Status         *models.StatusAgendamento `json:"status"`
    DataHoraInicio *time.Time                `json:"dataHoraInicio"`
    DataHoraFim    *time.Time                `json:"dataHoraFim"`
}

type UpdateStatusInput struct {
    Status models.StatusAgendamento `json:"status" binding:"required"`
}

// AgendamentoListDTO representa os dados essenciais para exibição em lista (Zero-Allocation de metadados inúteis).
type AgendamentoListDTO struct {
	ID             uint                     `json:"id"`
	DataHoraInicio time.Time                `json:"dataHoraInicio"`
	Status         models.StatusAgendamento `json:"status"`
	Motivo         string                   `json:"motivo"`
	PacienteNome   string                   `json:"pacienteNome"`
	DentistaNome   string                   `json:"dentistaNome"`
}

// Create godoc
// @Summary      Create an appointment
// @Description  Schedule a new appointment for a patient in the user's clinic
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      CreateAgendamentoInput  true  "Appointment Info"
// @Success      201    {object}  models.Agendamento
// @Router       /agendamentos [post]
func (h *AgendamentoHandler) Create(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input CreateAgendamentoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	var userID uint
	if userIDVal != nil {
		userID = userIDVal.(uint)
	} else {
		// Fallback se não vier do token, usa o que vier do payload
		userID = input.UsuarioCriacaoID
	}

	// TODO: Validate time conflicts

	agendamento := models.Agendamento{
		PacienteID:       input.PacienteID,
		DentistaID:       input.DentistaID,
		ClinicaID:        userClinicaID, // Force current clinic ID
		DataHoraInicio:   input.DataHoraInicio,
		DataHoraFim:      input.DataHoraFim,
		Motivo:           input.Motivo,
		UsuarioCriacaoID: userID,
		Status:           models.StatusAgendamentoAgendado,
	}

	if err := h.DB.Create(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agendamento"})
		return
	}

	c.JSON(http.StatusCreated, agendamento)
}

// FindAll godoc
// @Summary      List appointments
// @Description  List appointments with date and dentist filtering (clinic-scoped) and Pagination
// @Tags         agendamentos
// @Produce      json
// @Security     BearerAuth
// @Param        dataInicio  query     string  false  "Start Date (YYYY-MM-DD)"
// @Param        dataFim     query     string  false  "End Date (YYYY-MM-DD)"
// @Param        dentistaId  query     string  false  "Dentist ID"
// @Param        page        query     int     false  "Page number (default 1)"
// @Param        limit       query     int     false  "Limit per page (default 20, max 50)"
// @Success      200         {object}  utils.PaginatedResponse
// @Router       /agendamentos [get]
func (h *AgendamentoHandler) FindAll(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	reqClinicaID := c.Query("clinicaId")
	dataInicioStr := c.Query("dataInicio")
	dataFimStr := c.Query("dataFim")
	dentistaID := c.Query("dentistaId")
	profissionaisIDsStr := c.Query("profissionais_ids")

	// Fallback para data de início: se vazio, usa hoje para restringir a busca ao dia atual
	if dataInicioStr == "" {
		dataInicioStr = time.Now().Format("2006-01-02")
	}

	// Se dataFim estiver vazio, trava a busca exatamente para o final do mesmo dia (evitando buscar todo o futuro)
	if dataFimStr == "" {
		dataFimStr = dataInicioStr
	}

	page, limit := utils.GetPaginationParams(c)

	query := h.DB.Model(&models.Agendamento{})

	if userRole == "ADMIN_TOTAL" && reqClinicaID != "" {
		query = query.Where("agendamentos.clinica_id = ?", reqClinicaID)
	} else {
		query = query.Where("agendamentos.clinica_id = ?", userClinicaID)
	}

	if profissionaisIDsStr != "" {
		ids := strings.Split(profissionaisIDsStr, ",")
		query = query.Where("agendamentos.dentista_id IN ?", ids)
	} else if dentistaID != "" {
		query = query.Where("agendamentos.dentista_id = ?", dentistaID)
	}

	// Aplica a janela de tempo rigorosa de exatamente 1 dia (ou o range fornecido)
	query = query.Where("agendamentos.data_hora_inicio >= ? AND agendamentos.data_hora_inicio <= ?", dataInicioStr+" 00:00:00", dataFimStr+" 23:59:59")

	// 1. Fazer o Count Total antes do OFFSET/LIMIT para os metadados da resposta
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count agendamentos"})
		return
	}

	// 2. Fazer a projeção (Select/Joins) sem preloads
	var dtos []AgendamentoListDTO
	query = query.
		Select("agendamentos.id", "agendamentos.data_hora_inicio", "agendamentos.status", "agendamentos.motivo", "pacientes.nome as paciente_nome", "users.nome as dentista_nome").
		Joins("left join pacientes on pacientes.id = agendamentos.paciente_id").
		Joins("left join dentistas on dentistas.id = agendamentos.dentista_id").
		Joins("left join users on users.id = dentistas.usuario_id")

	// 3. Aplicar ordenação e a paginação nativa (OFFSET/LIMIT) via Utility Scope
	if err := query.Order("agendamentos.data_hora_inicio ASC").Scopes(utils.Paginate(page, limit)).Scan(&dtos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agendamentos"})
		return
	}

	// Evitar retornar 'null' no json quando slice estiver vazio (cria um slice vazio [] em vez de nulo)
	if dtos == nil {
		dtos = make([]AgendamentoListDTO, 0)
	}

	// Retornar a estrutura envelopada
	response := utils.PaginatedResponse{
		Data: dtos,
		Meta: utils.PaginationMeta{
			CurrentPage:  page,
			Limit:        limit,
			TotalRecords: totalRecords,
		},
	}

	c.JSON(http.StatusOK, response)
}

// FindOne godoc
// @Summary      Get appointment details
// @Description  Get details of a specific appointment (clinic-scoped)
// @Tags         agendamentos
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Appointment ID"
// @Success      200  {object}  models.Agendamento
// @Failure      403  {object}  map[string]string
// @Router       /agendamentos/{id} [get]
func (h *AgendamentoHandler) FindOne(c *gin.Context) {
    userClinicaID, err := getClinicaIDFromContext(c)
    userRole := getUserRoleFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
        return
    }

    id := c.Param("id")
    var agendamento models.Agendamento
    if err := h.DB.Preload("Paciente").Preload("Dentista").First(&agendamento, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
        return
    }

    if userRole != "ADMIN_TOTAL" && agendamento.ClinicaID != userClinicaID {
        c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
        return
    }

    c.JSON(http.StatusOK, agendamento)
}

// Update godoc
// @Summary      Update appointment
// @Description  Update details of an appointment (e.g., MotivoConsulta, Status)
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Appointment ID"
// @Param        input  body      UpdateAgendamentoInput  true  "Updated Fields"
// @Success      200    {object}  models.Agendamento
// @Router       /agendamentos/{id} [patch]
func (h *AgendamentoHandler) Update(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	var input UpdateAgendamentoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agendamento models.Agendamento
	if err := h.DB.First(&agendamento, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if userRole != "ADMIN_TOTAL" && agendamento.ClinicaID != userClinicaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if input.MotivoConsulta != nil {
		agendamento.Motivo = *input.MotivoConsulta
	}
	if input.Status != nil {
		agendamento.Status = *input.Status
	}
	if input.DataHoraInicio != nil {
		agendamento.DataHoraInicio = *input.DataHoraInicio
	}
	if input.DataHoraFim != nil {
		agendamento.DataHoraFim = *input.DataHoraFim
	}

	if err := h.DB.Save(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agendamento"})
		return
	}

	c.JSON(http.StatusOK, agendamento)
}

// UpdateStatus godoc
// @Summary      Update appointment status
// @Description  Change the status of an appointment (e.g., Confirmed, Cancelled, Completed)
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Appointment ID"
// @Param        input  body      UpdateStatusInput  true  "New Status"
// @Success      200    {object}  models.Agendamento
// @Router       /agendamentos/{id}/status [patch]
func (h *AgendamentoHandler) UpdateStatus(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")
	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agendamento models.Agendamento
	if err := h.DB.First(&agendamento, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if userRole != "ADMIN_TOTAL" && agendamento.ClinicaID != userClinicaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	agendamento.Status = input.Status
	if err := h.DB.Save(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, agendamento)
}

// Delete godoc
// @Summary      Delete appointment
// @Description  Delete an appointment from the schedule
// @Tags         agendamentos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string             true  "Appointment ID"
// @Success      204    "No Content"
// @Router       /agendamentos/{id} [delete]
func (h *AgendamentoHandler) Delete(c *gin.Context) {
	userClinicaID, err := getClinicaIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	id := c.Param("id")

	var agendamento models.Agendamento
	if err := h.DB.First(&agendamento, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if userRole != "ADMIN_TOTAL" && agendamento.ClinicaID != userClinicaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento not found"})
		return
	}

	if err := h.DB.Delete(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agendamento"})
		return
	}

	c.Status(http.StatusNoContent)
}
