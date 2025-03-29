package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"rating-system/internal/domain/model"
	"rating-system/pkg/pagination"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateRating(ctx context.Context, rating *model.Rating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockRepository) GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockRepository) GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error) {
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockRepository) GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error) {
	args := m.Called(ctx, serviceID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Rating), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateRating(ctx context.Context, rating *model.Rating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockRepository) CalculateAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AverageRating), args.Error(1)
}

func (m *MockRepository) CreateReview(ctx context.Context, review *model.Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockRepository) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReviewWithRating), args.Error(1)
}

func (m *MockRepository) GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error) {
	args := m.Called(ctx, serviceID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.ReviewWithRating), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateReview(ctx context.Context, review *model.Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockRepository) GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error) {
	args := m.Called(ctx, reviewID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Comment), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateComment(ctx context.Context, comment *model.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func TestCreateRating(t *testing.T) {
	logger := logrus.New()
	repo := new(MockRepository)
	service := NewRatingService(repo, logger)
	ctx := context.Background()

	userID := uuid.New()
	serviceID := uuid.New()
	score := 4

	// Test case 1: User has not previously rated this service
	repo.On("GetRatingByUserAndService", ctx, userID, serviceID).
		Return(nil, errors.New("not found")).Once()
	
	repo.On("CreateRating", ctx, mock.MatchedBy(func(r *model.Rating) bool {
		return r.UserID == userID && r.ServiceID == serviceID && r.Score == score
	})).Return(nil).Once()

	rating, err := service.CreateRating(ctx, userID, serviceID, score)
	assert.NoError(t, err)
	assert.NotNil(t, rating)
	assert.Equal(t, userID, rating.UserID)
	assert.Equal(t, serviceID, rating.ServiceID)
	assert.Equal(t, score, rating.Score)

	// Test case 2: User has already rated this service
	existingRating, _ := model.NewRating(userID, serviceID, 3)
	repo.On("GetRatingByUserAndService", ctx, userID, serviceID).
		Return(existingRating, nil).Once()
	
	repo.On("UpdateRating", ctx, mock.MatchedBy(func(r *model.Rating) bool {
		return r.ID == existingRating.ID && r.Score == score
	})).Return(nil).Once()

	updatedRating, err := service.CreateRating(ctx, userID, serviceID, score)
	assert.NoError(t, err)
	assert.NotNil(t, updatedRating)
	assert.Equal(t, existingRating.ID, updatedRating.ID)
	assert.Equal(t, score, updatedRating.Score)

	repo.AssertExpectations(t)
}

func TestGetAverageRating(t *testing.T) {
	logger := logrus.New()
	repo := new(MockRepository)
	service := NewRatingService(repo, logger)
	ctx := context.Background()

	serviceID := uuid.New()
	expected := &model.AverageRating{
		ServiceID:    serviceID,
		AverageScore: 4.5,
		TotalRatings: 10,
	}

	repo.On("CalculateAverageRating", ctx, serviceID).Return(expected, nil).Once()

	average, err := service.GetAverageRating(ctx, serviceID)
	assert.NoError(t, err)
	assert.Equal(t, expected, average)

	repo.AssertExpectations(t)
}

func TestCreateReview(t *testing.T) {
	logger := logrus.New()
	repo := new(MockRepository)
	service := NewRatingService(repo, logger)
	ctx := context.Background()

	userID := uuid.New()
	serviceID := uuid.New()
	ratingID := uuid.New()
	title := "Great service"
	content := "This service was really helpful"

	// Setup mocks
	rating, _ := model.NewRating(userID, serviceID, 5)
	rating.ID = ratingID

	repo.On("GetRatingByID", ctx, ratingID).Return(rating, nil).Once()
	
	repo.On("CreateReview", ctx, mock.MatchedBy(func(r *model.Review) bool {
		return r.UserID == userID && r.ServiceID == serviceID && 
		       r.RatingID == ratingID && r.Title == title && r.Content == content
	})).Return(nil).Once()

	review, err := service.CreateReview(ctx, userID, serviceID, ratingID, title, content)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, userID, review.UserID)
	assert.Equal(t, serviceID, review.ServiceID)
	assert.Equal(t, ratingID, review.RatingID)
	assert.Equal(t, title, review.Title)
	assert.Equal(t, content, review.Content)

	repo.AssertExpectations(t)
}

func TestCreateComment(t *testing.T) {
	logger := logrus.New()
	repo := new(MockRepository)
	service := NewRatingService(repo, logger)
	ctx := context.Background()

	userID := uuid.New()
	reviewID := uuid.New()
	content := "This is a comment on the review"

	// Setup mocks
	reviewWithRating := &model.ReviewWithRating{
		Review: model.Review{
			ID:        reviewID,
			UserID:    uuid.New(),
			ServiceID: uuid.New(),
			RatingID:  uuid.New(),
			Title:     "Some Title",
			Content:   "Some Content",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Score: 4,
	}

	repo.On("GetReviewByID", ctx, reviewID).Return(reviewWithRating, nil).Once()
	
	repo.On("CreateComment", ctx, mock.MatchedBy(func(c *model.Comment) bool {
		return c.UserID == userID && c.ReviewID == reviewID && c.Content == content
	})).Return(nil).Once()

	comment, err := service.CreateComment(ctx, userID, reviewID, content)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, userID, comment.UserID)
	assert.Equal(t, reviewID, comment.ReviewID)
	assert.Equal(t, content, comment.Content)

	repo.AssertExpectations(t)
}
