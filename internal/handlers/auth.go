package handlers

import (
	"net/http"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
    DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
    return &AuthHandler{DB: db}
}

type RegisterInput struct {
    Nome        string `json:"nome" binding:"required"`
    Email       string `json:"email" binding:"required,email"`
    Senha       string `json:"senha" binding:"required,min=6"`
    Telefone    string `json:"telefone"`
    CPF         string `json:"cpf"`
    TipoUsuario string `json:"tipoUsuario" binding:"required"`
}

type LoginInput struct {
    Email string `json:"email" binding:"required,email"`
    Senha string `json:"senha" binding:"required"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account in the system
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      RegisterInput  true  "Registration Info"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Router       /auth/signup [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var count int64
	h.DB.Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Nome:        input.Nome,
		Email:       input.Email,
		SenhaHash:   string(hashedPassword),
		Telefone:    input.Telefone,
		CPF:         input.CPF,
		TipoUsuario: input.TipoUsuario,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      LoginInput  true  "Login Credentials"
// @Success      200    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := h.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(input.Senha)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate JWT
    cfg, _ := config.LoadConfig()
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":         user.ID,
        "name":        user.Nome,
        "email":       user.Email,
        "tipoUsuario": user.TipoUsuario,
        "clinicaId":   user.ClinicaID,
        "exp":         time.Now().Add(time.Hour * 24).Unix(),
    })

    tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Me godoc
// @Summary      Get current user
// @Description  Retrieve the profile of the currently authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
