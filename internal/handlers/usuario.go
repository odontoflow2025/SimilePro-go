package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UsuarioHandler struct {
    DB *gorm.DB
}

func NewUsuarioHandler(db *gorm.DB) *UsuarioHandler {
    return &UsuarioHandler{DB: db}
}

// FindAll godoc
// @Summary      List users
// @Description  Retrieve a list of all users in the user's clinic
// @Tags         usuarios
// @Produce      json
// @Security     BearerAuth
// @Router       /usuarios [get]
func (h *UsuarioHandler) FindAll(c *gin.Context) {
	userClinicaID, exists := c.Get("clinicaID")
	userRole, _ := c.Get("userRole")
	
	if !exists || userClinicaID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Clinic not identified"})
		return
	}

	query := h.DB.Model(&models.User{})

	if userRole == "ADMIN_TOTAL" {
		// Can see all clinics, or filter by a specific one
		reqClinicaID := c.Query("clinicaId")
		if reqClinicaID != "" {
			query = query.Where("clinica_id = ?", reqClinicaID)
		}
	} else {
		// Restricted to their own clinic
		query = query.Where("clinica_id = ?", userClinicaID)
	}

	var usuarios []models.User
	if err := query.Find(&usuarios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usuarios"})
		return
	}

	c.JSON(http.StatusOK, usuarios)
}

// Update godoc
// @Summary      Update user
// @Description  Update user's basic info
// @Tags         usuarios
// @Produce      json
// @Security     BearerAuth
// @Router       /usuarios/{id} [patch]
func (h *UsuarioHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Nome     string `json:"nome"`
		Telefone string `json:"telefone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.DB.Model(&user).Updates(models.User{Nome: input.Nome, Telefone: input.Telefone})
	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary      Delete user
// @Description  Delete user and potentially their dependencies
// @Tags         usuarios
// @Produce      json
// @Security     BearerAuth
// @Router       /usuarios/{id} [delete]
func (h *UsuarioHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent deleting ADMIN_TOTAL
	if user.TipoUsuario == "ADMIN_TOTAL" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete root admin"})
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
