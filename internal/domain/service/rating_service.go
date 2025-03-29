package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	
	"rating-system/internal/domain/model"
	"rating-system/internal/domain/port"
	"rating-system/pkg/pagination"
)

// RatingService implements the Service port
type RatingService struct {
	repo port.Repository
	log  *logrus.Logger
}

// NewRatingService creates a new rating service
func NewRatingService(repo port.Repository, log *logrus.Logger) port.Service {
	return &RatingService{
		repo: repo,
		log:  log,
	}
}

// CreateRating creates a new rating
func (s *RatingService) CreateRating(ctx context.Context, userID, serviceID uuid.UUID, score int) (*model.Rating, error) {
	// Check if user already rated this service
	existingRating, err := s.repo.GetRatingByUserAndService(ctx, userID, serviceID)
	if err == nil && existingRating != nil {
		// Update existing rating instead of creating a new one
		existingRating.UpdateScore(score)
		if err := s.repo.UpdateRating(ctx, existingRating); err != nil {
			s.log.WithError(err).Error("Failed to update existing rating")
			return nil, err
		}
		return existingRating, nil
	}

	// Create new rating
	rating, err := model.NewRating(userID, serviceID, score)
	if err != nil {
		s.log.WithError(err).Error("Failed to create rating model")
		return nil, err
	}

	if err := s.repo.CreateRating(ctx, rating); err != nil {
		s.log.WithError(err).Error("Failed to create rating in repository")
		return nil, err
	}

	return rating, nil
}

// GetRatingByID retrieves a rating by ID
func (s *RatingService) GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error) {
	rating, err := s.repo.GetRatingByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get rating by ID")
		return nil, err
	}
	return rating, nil
}

// GetRatingByUserAndService retrieves a rating by user and service
func (s *RatingService) GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error) {
	rating, err := s.repo.GetRatingByUserAndService(ctx, userID, serviceID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get rating by user and service")
		return nil, err
	}
	return rating, nil
}

// GetRatingsByService retrieves ratings by service ID with pagination
func (s *RatingService) GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error) {
	ratings, total, err := s.repo.GetRatingsByService(ctx, serviceID, params)
	if err != nil {
		s.log.WithError(err).Error("Failed to get ratings by service")
		return nil, 0, err
	}
	return ratings, total, nil
}

// UpdateRating updates an existing rating
func (s *RatingService) UpdateRating(ctx context.Context, id uuid.UUID, score int) (*model.Rating, error) {
	rating, err := s.repo.GetRatingByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get rating for update")
		return nil, err
	}

	if err := rating.UpdateScore(score); err != nil {
		s.log.WithError(err).Error("Failed to update rating score")
		return nil, err
	}

	if err := s.repo.UpdateRating(ctx, rating); err != nil {
		s.log.WithError(err).Error("Failed to update rating in repository")
		return nil, err
	}

	return rating, nil
}

// GetAverageRating calculates the average rating for a service
func (s *RatingService) GetAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error) {
	average, err := s.repo.CalculateAverageRating(ctx, serviceID)
	if err != nil {
		s.log.WithError(err).Error("Failed to calculate average rating")
		return nil, err
	}
	return average, nil
}

// CreateReview creates a new review
func (s *RatingService) CreateReview(ctx context.Context, userID, serviceID uuid.UUID, ratingID uuid.UUID, title, content string) (*model.Review, error) {
	// Validate that rating exists and belongs to the user and service
	rating, err := s.repo.GetRatingByID(ctx, ratingID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get rating for review creation")
		return nil, errors.New("rating not found")
	}

	if rating.UserID != userID || rating.ServiceID != serviceID {
		s.log.Error("Rating doesn't match user or service")
		return nil, errors.New("rating doesn't match user or service")
	}

	review, err := model.NewReview(userID, serviceID, ratingID, title, content)
	if err != nil {
		s.log.WithError(err).Error("Failed to create review model")
		return nil, err
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		s.log.WithError(err).Error("Failed to create review in repository")
		return nil, err
	}

	return review, nil
}

// GetReviewByID retrieves a review by ID
func (s *RatingService) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error) {
	review, err := s.repo.GetReviewByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get review by ID")
		return nil, err
	}
	return review, nil
}

// GetReviewsByService retrieves reviews by service ID with pagination
func (s *RatingService) GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error) {
	reviews, total, err := s.repo.GetReviewsByService(ctx, serviceID, params)
	if err != nil {
		s.log.WithError(err).Error("Failed to get reviews by service")
		return nil, 0, err
	}
	return reviews, total, nil
}

// UpdateReview updates an existing review
func (s *RatingService) UpdateReview(ctx context.Context, id uuid.UUID, title, content string) (*model.Review, error) {
	// Get the review with rating to ensure it exists
	reviewWithRating, err := s.repo.GetReviewByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get review for update")
		return nil, err
	}

	// Convert to regular review
	review := &model.Review{
		ID:        reviewWithRating.ID,
		UserID:    reviewWithRating.UserID,
		ServiceID: reviewWithRating.ServiceID,
		RatingID:  reviewWithRating.RatingID,
		Title:     reviewWithRating.Title,
		Content:   reviewWithRating.Content,
		CreatedAt: reviewWithRating.CreatedAt,
		UpdatedAt: reviewWithRating.UpdatedAt,
	}

	if err := review.UpdateContent(title, content); err != nil {
		s.log.WithError(err).Error("Failed to update review content")
		return nil, err
	}

	if err := s.repo.UpdateReview(ctx, review); err != nil {
		s.log.WithError(err).Error("Failed to update review in repository")
		return nil, err
	}

	return review, nil
}

// CreateComment creates a new comment
func (s *RatingService) CreateComment(ctx context.Context, userID, reviewID uuid.UUID, content string) (*model.Comment, error) {
	// Verify that review exists
	_, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get review for comment creation")
		return nil, errors.New("review not found")
	}

	comment, err := model.NewComment(userID, reviewID, content)
	if err != nil {
		s.log.WithError(err).Error("Failed to create comment model")
		return nil, err
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		s.log.WithError(err).Error("Failed to create comment in repository")
		return nil, err
	}

	return comment, nil
}

// GetCommentByID retrieves a comment by ID
func (s *RatingService) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	comment, err := s.repo.GetCommentByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get comment by ID")
		return nil, err
	}
	return comment, nil
}

// GetCommentsByReview retrieves comments by review ID with pagination
func (s *RatingService) GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error) {
	comments, total, err := s.repo.GetCommentsByReview(ctx, reviewID, params)
	if err != nil {
		s.log.WithError(err).Error("Failed to get comments by review")
		return nil, 0, err
	}
	return comments, total, nil
}

// UpdateComment updates an existing comment
func (s *RatingService) UpdateComment(ctx context.Context, id uuid.UUID, content string) (*model.Comment, error) {
	comment, err := s.repo.GetCommentByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get comment for update")
		return nil, err
	}

	if err := comment.UpdateContent(content); err != nil {
		s.log.WithError(err).Error("Failed to update comment content")
		return nil, err
	}

	if err := s.repo.UpdateComment(ctx, comment); err != nil {
		s.log.WithError(err).Error("Failed to update comment in repository")
		return nil, err
	}

	return comment, nil
}
