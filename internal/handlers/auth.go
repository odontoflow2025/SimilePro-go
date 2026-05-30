package handlers

import (
	"net/http"
	"odonto-flow-go/internal/models"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"context"
	"odonto-flow-go/internal/database"
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
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
        "jti":         uuid.New().String(),
        "sub":         user.ID,
        "name":        user.Nome,
        "email":       user.Email,
        "tipoUsuario": user.TipoUsuario,
        "clinicaId":   user.ClinicaID,
        "exp":         time.Now().Add(time.Hour * 24).Unix(),
    })

    privateKeyBytes, err := os.ReadFile("private.pem")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read private key"})
        return
    }
    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse private key"})
        return
    }

    tokenString, err := token.SignedString(privateKey)
    if (err != nil) {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // Set JWT as HttpOnly Cookie
    maxAge := 86400 
    
    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "auth_token",
        Value:    tokenString,
        MaxAge:   maxAge,
        Path:     "/",
        Domain:   "",
        Secure:   true, // SameSiteNone requires Secure: true
        HttpOnly: true,
        SameSite: http.SameSiteNoneMode, // Allows cross-origin cookies
    })

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
	// Extrair o token do cookie (ou header)
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie != "" {
		// Parseamos ignorando a assinatura apenas para extrair as claims publicas
		token, _, _ := new(jwt.Parser).ParseUnverified(cookie, jwt.MapClaims{})
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			jti, okJTI := claims["jti"].(string)
			expFloat, okExp := claims["exp"].(float64)
			
			if okJTI && okExp {
				expTime := time.Unix(int64(expFloat), 0)
				timeRemaining := time.Until(expTime)
				if timeRemaining > 0 {
					ctx := context.Background()
					if database.RedisClient != nil {
						database.RedisClient.Set(ctx, "blacklist:"+jti, "revogado", timeRemaining)
					}
				}
			}
		}
	}

	// Clear the cookie by setting maxAge to -1
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
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
