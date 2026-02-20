package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
    DB *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
    return &AdminHandler{DB: db}
}

// --- Stats ---

// GetStats godoc
// @Summary      System global statistics
// @Description  Get counts of patients, dentists, and consultations across the system
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	var pacientes int64
	h.DB.Model(&models.Paciente{}).Count(&pacientes)

	var dentistas int64
	h.DB.Model(&models.Dentista{}).Count(&dentistas)

	var consultas int64
	h.DB.Model(&models.Agendamento{}).Count(&consultas) // Total appointments

	stats := gin.H{
		"totalPacientes": pacientes,
		"totalDentistas": dentistas,
		"totalConsultas": consultas,
		"ocupacao":       0, // Mock for now
		"novosPacientes": 0, // Mock for now
	}

	c.JSON(http.StatusOK, stats)
}

// --- Configurações ---

// ListConfigs godoc
// @Summary      List system settings
// @Description  Retrieve all system-wide configurations and their values
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.ConfiguracaoSistema
// @Router       /admin/config [get]
func (h *AdminHandler) ListConfigs(c *gin.Context) {
	var configs []models.ConfiguracaoSistema
	if err := h.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configs"})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// GetConfig godoc
// @Summary      Get specific setting
// @Description  Retrieve the value and description of a configuration key
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        chave  path      string  true  "Config Key"
// @Success      200    {object}  models.ConfiguracaoSistema
// @Failure      404    {object}  map[string]string
// @Router       /admin/config/{chave} [get]
func (h *AdminHandler) GetConfig(c *gin.Context) {
	key := c.Param("chave")
	var config models.ConfiguracaoSistema
	if err := h.DB.Where("chave = ?", key).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}
	c.JSON(http.StatusOK, config) // Returns full object, frontend might want {valor: ...}
}

// SetConfig godoc
// @Summary      Update or create system setting
// @Description  Create a new configuration key or update an existing one's value and description
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chave  path      string  true  "Config Key"
// @Param        input  body      object  true  "New Value & Description"
// @Success      200    {object}  models.ConfiguracaoSistema
// @Router       /admin/config/{chave} [put]
func (h *AdminHandler) SetConfig(c *gin.Context) {
    key := c.Param("chave")
    var input struct {
        Valor     string `json:"valor"`
        Descricao string `json:"descricao"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get userID from context
    userIDVal, _ := c.Get("userID")
    var userID uint
    switch v := userIDVal.(type) {
    case float64:
        userID = uint(v)
    case uint:
        userID = v
    case int:
        userID = uint(v)
    }

    // Upsert
    var config models.ConfiguracaoSistema
    err := h.DB.Where("chave = ?", key).First(&config).Error
    
    if err == nil {
        // Update
        config.Valor = input.Valor
        if input.Descricao != "" {
            config.Descricao = input.Descricao
        }
        config.UsuarioAtualizacaoID = &userID
        h.DB.Save(&config)
    } else {
        // Create
        config = models.ConfiguracaoSistema{
            Chave:                key,
            Valor:                input.Valor,
            Descricao:            input.Descricao,
            UsuarioAtualizacaoID: &userID,
        }
        h.DB.Create(&config)
    }

    c.JSON(http.StatusOK, config)
}

// --- Logs ---

// GetLogs godoc
// @Summary      System audit logs
// @Description  Retrieve system audit logs with filtering capability
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        dataInicio query     string  false  "Start Date (ISO)"
// @Param        dataFim    query     string  false  "End Date (ISO)"
// @Param        usuarioId  query     string  false  "Filter by User ID"
// @Param        modulo     query     string  false  "Filter by Module"
// @Success      200        {array}   models.AuditLog
// @Router       /admin/logs [get]
func (h *AdminHandler) GetLogs(c *gin.Context) {
	// Filters: date range, user, module/resource
	dataInicio := c.Query("dataInicio")
	dataFim := c.Query("dataFim")
	usuarioID := c.Query("usuarioId")
	recurso := c.Query("modulo") // Maps to Recurso
	limit := 50

	query := h.DB.Model(&models.AuditLog{}).Preload("Usuario").Order("created_at DESC")

	if dataInicio != "" && dataFim != "" {
		// Parse dates if necessary, or pass strings if GORM handles it.
		// Better to parse to ensure safety.
		// Assuming ISO format
		start, err1 := time.Parse(time.RFC3339, dataInicio)
		end, err2 := time.Parse(time.RFC3339, dataFim)
		if err1 == nil && err2 == nil {
			query = query.Where("created_at BETWEEN ? AND ?", start, end)
		}
	}

	if usuarioID != "" {
		query = query.Where("usuario_id = ?", usuarioID)
	}

	if recurso != "" {
		query = query.Where("recurso = ?", recurso)
	}

	var logs []models.AuditLog
	query.Limit(limit).Find(&logs)

	c.JSON(http.StatusOK, logs)
}
