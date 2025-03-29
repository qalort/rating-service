package port

import (
	"context"

	"github.com/google/uuid"

	"rating-system/internal/domain/model"
)

// AuthService defines the interface for authentication services
type AuthService interface {
	// Register registers a new user
	Register(ctx context.Context, username, email, password string) (*model.UserResponse, string, error)
	
	// Login authenticates a user and returns a token
	Login(ctx context.Context, email, password string) (*model.UserResponse, string, error)
	
	// ValidateToken validates a token and returns the user ID if valid
	ValidateToken(token string) (uuid.UUID, error)
}