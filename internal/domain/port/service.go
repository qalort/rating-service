package port

import (
	"context"

	"github.com/google/uuid"
	
	"rating-system/internal/domain/model"
	"rating-system/pkg/pagination"
)

// Service defines the port for service operations
type Service interface {
	// Rating operations
	CreateRating(ctx context.Context, userID, serviceID uuid.UUID, score int) (*model.Rating, error)
	GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error)
	GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error)
	GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error)
	UpdateRating(ctx context.Context, id uuid.UUID, score int) (*model.Rating, error)
	GetAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error)
	
	// Review operations
	CreateReview(ctx context.Context, userID, serviceID uuid.UUID, ratingID uuid.UUID, title, content string) (*model.Review, error)
	GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error)
	GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error)
	UpdateReview(ctx context.Context, id uuid.UUID, title, content string) (*model.Review, error)
	
	// Comment operations
	CreateComment(ctx context.Context, userID, reviewID uuid.UUID, content string) (*model.Comment, error)
	GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error)
	UpdateComment(ctx context.Context, id uuid.UUID, content string) (*model.Comment, error)
}
