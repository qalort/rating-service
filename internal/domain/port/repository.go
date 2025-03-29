package port

import (
        "context"

        "github.com/google/uuid"
        
        "rating-system/internal/domain/model"
        "rating-system/pkg/pagination"
)

// Repository defines the port for repository operations
type Repository interface {
        // User operations
        CreateUser(ctx context.Context, user *model.User) error
        GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
        GetUserByEmail(ctx context.Context, email string) (*model.User, error)
        GetUserByUsername(ctx context.Context, username string) (*model.User, error)

        // Rating operations
        CreateRating(ctx context.Context, rating *model.Rating) error
        GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error)
        GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error)
        GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error)
        UpdateRating(ctx context.Context, rating *model.Rating) error
        CalculateAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error)
        
        // Review operations
        CreateReview(ctx context.Context, review *model.Review) error
        GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error)
        GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error)
        UpdateReview(ctx context.Context, review *model.Review) error
        
        // Comment operations
        CreateComment(ctx context.Context, comment *model.Comment) error
        GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
        GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error)
        UpdateComment(ctx context.Context, comment *model.Comment) error
}
