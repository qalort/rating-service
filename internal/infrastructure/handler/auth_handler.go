package handler

import (
        "net/http"
        "strings"

        "github.com/gin-gonic/gin"
        "github.com/sirupsen/logrus"

        "rating-system/internal/domain/port"
        "rating-system/internal/service"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
        authService port.AuthService
        log         *logrus.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService port.AuthService, log *logrus.Logger) *AuthHandler {
        return &AuthHandler{
                authService: authService,
                log:         log,
        }
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
        Username string `json:"username" binding:"required"`
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
        Token string      `json:"token"`
        User  interface{} `json:"user"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        user, token, err := h.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
        if err != nil {
                if err == service.ErrUserAlreadyExists {
                        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
                        return
                }
                h.log.WithError(err).Error("Failed to register user")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
                return
        }

        c.JSON(http.StatusCreated, AuthResponse{
                Token: token,
                User:  user,
        })
}

// Login handles user login
// @Summary Login a user
// @Description Authenticate a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
        if err != nil {
                if err == service.ErrInvalidCredentials {
                        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
                        return
                }
                h.log.WithError(err).Error("Failed to login user")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
                return
        }

        c.JSON(http.StatusOK, AuthResponse{
                Token: token,
                User:  user,
        })
}

// AuthMiddleware is a middleware to authenticate requests
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                // Get authorization header
                authHeader := c.GetHeader("Authorization")
                if authHeader == "" {
                        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
                        c.Abort()
                        return
                }

                // Check Bearer format
                parts := strings.Split(authHeader, " ")
                if len(parts) != 2 || parts[0] != "Bearer" {
                        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
                        c.Abort()
                        return
                }

                // Validate token
                userID, err := h.authService.ValidateToken(parts[1])
                if err != nil {
                        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
                        c.Abort()
                        return
                }

                // Set user ID in context
                c.Set("userID", userID)
                c.Next()
        }
}