package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"rating-system/internal/domain/model"
	"rating-system/internal/domain/port"
	"rating-system/internal/infrastructure/auth"
)

// Errors related to authentication
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService implements the AuthService interface
type AuthService struct {
	repository port.Repository
	jwtService *auth.JWTService
	log        *logrus.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(repository port.Repository, log *logrus.Logger) (port.AuthService, error) {
	jwtService, err := auth.NewJWTService()
	if err != nil {
		return nil, err
	}

	return &AuthService{
		repository: repository,
		jwtService: jwtService,
		log:        log,
	}, nil
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*model.UserResponse, string, error) {
	// Check if user already exists with the same email or username
	existingUser, err := s.repository.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", ErrUserAlreadyExists
	}

	existingUser, err = s.repository.GetUserByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, "", ErrUserAlreadyExists
	}

	// Create new user
	user, err := model.NewUser(username, email, password)
	if err != nil {
		return nil, "", err
	}

	// Save user to database
	err = s.repository.CreateUser(ctx, user)
	if err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	response := user.ToResponse()
	return &response, token, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.UserResponse, string, error) {
	// Get user by email
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	response := user.ToResponse()
	return &response, token, nil
}

// ValidateToken validates a token and returns the user ID if valid
func (s *AuthService) ValidateToken(token string) (uuid.UUID, error) {
	userID, err := s.jwtService.ExtractUserID(token)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}