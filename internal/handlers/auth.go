package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
    DB        *gorm.DB
    JWTSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
    return &AuthHandler{DB: db, JWTSecret: jwtSecret}
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

	// Security check for role
	tipoUsuario := input.TipoUsuario
	// For public registration, block internal admin roles
	if tipoUsuario == "ADMIN_TOTAL" || tipoUsuario == "ADMIN_GERENCIAL" || tipoUsuario == "" {
		tipoUsuario = "RECEPCIONISTA" // Default to lower privilege
	}

	user := models.User{
		Nome:        input.Nome,
		Email:       input.Email,
		SenhaHash:   string(hashedPassword),
		Telefone:    input.Telefone,
		CPF:         input.CPF,
		TipoUsuario: tipoUsuario,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
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
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":         user.ID,
        "name":        user.Nome,
        "email":       user.Email,
        "tipoUsuario": user.TipoUsuario,
        "clinicaId":   user.ClinicaID,
        "exp":         time.Now().Add(time.Hour * 24).Unix(),
    })

    tokenString, err := token.SignedString([]byte(h.JWTSecret))
    if (err != nil) {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // Set JWT as HttpOnly Cookie
    // MaxAge in seconds (24h)
    maxAge := 86400 
    
    // In production, Secure should be true. 
    // Here we use false for local development compatibility unless otherwise configured.
    c.SetCookie("auth_token", tokenString, maxAge, "/", "", false, true)

    c.JSON(http.StatusOK, gin.H{
        "token": tokenString, // Keep for legacy compatibility if needed
        "message": "Login successful",
    })
}

// Logout godoc
// @Summary      Logout user
// @Description  Clear the authentication cookie
// @Tags         auth
// @Produce      json
// @Success      200    {object}  map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the cookie by setting maxAge to -1
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
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
