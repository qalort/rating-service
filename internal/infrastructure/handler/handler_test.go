package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"rating-system/internal/domain/model"
	"rating-system/pkg/pagination"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateRating(ctx interface{}, userID, serviceID uuid.UUID, score int) (*model.Rating, error) {
	args := m.Called(ctx, userID, serviceID, score)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockService) GetRatingByID(ctx interface{}, id uuid.UUID) (*model.Rating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockService) GetRatingByUserAndService(ctx interface{}, userID, serviceID uuid.UUID) (*model.Rating, error) {
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockService) GetRatingsByService(ctx interface{}, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error) {
	args := m.Called(ctx, serviceID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Rating), args.Int(1), args.Error(2)
}

func (m *MockService) UpdateRating(ctx interface{}, id uuid.UUID, score int) (*model.Rating, error) {
	args := m.Called(ctx, id, score)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

func (m *MockService) GetAverageRating(ctx interface{}, serviceID uuid.UUID) (*model.AverageRating, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AverageRating), args.Error(1)
}

func (m *MockService) CreateReview(ctx interface{}, userID, serviceID uuid.UUID, ratingID uuid.UUID, title, content string) (*model.Review, error) {
	args := m.Called(ctx, userID, serviceID, ratingID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Review), args.Error(1)
}

func (m *MockService) GetReviewByID(ctx interface{}, id uuid.UUID) (*model.ReviewWithRating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReviewWithRating), args.Error(1)
}

func (m *MockService) GetReviewsByService(ctx interface{}, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error) {
	args := m.Called(ctx, serviceID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.ReviewWithRating), args.Int(1), args.Error(2)
}

func (m *MockService) UpdateReview(ctx interface{}, id uuid.UUID, title, content string) (*model.Review, error) {
	args := m.Called(ctx, id, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Review), args.Error(1)
}

func (m *MockService) CreateComment(ctx interface{}, userID, reviewID uuid.UUID, content string) (*model.Comment, error) {
	args := m.Called(ctx, userID, reviewID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockService) GetCommentByID(ctx interface{}, id uuid.UUID) (*model.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockService) GetCommentsByReview(ctx interface{}, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error) {
	args := m.Called(ctx, reviewID, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Comment), args.Int(1), args.Error(2)
}

func (m *MockService) UpdateComment(ctx interface{}, id uuid.UUID, content string) (*model.Comment, error) {
	args := m.Called(ctx, id, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func setupTest() (*gin.Engine, *MockService) {
	gin.SetMode(gin.TestMode)
	
	router := gin.Default()
	mockService := new(MockService)
	
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	handler := NewHandler(mockService, logger)
	
	apiGroup := router.Group("/api/v1")
	
	// Set up routes
	ratings := apiGroup.Group("/ratings")
	ratings.POST("", handler.CreateRating)
	ratings.GET("/service/:serviceID", handler.GetRatingsByService)
	ratings.GET("/service/:serviceID/average", handler.GetAverageRating)
	ratings.GET("/user/:userID/service/:serviceID", handler.GetUserRating)
	
	reviews := apiGroup.Group("/reviews")
	reviews.POST("", handler.CreateReview)
	reviews.GET("/service/:serviceID", handler.GetReviewsByService)
	reviews.GET("/:reviewID", handler.GetReviewByID)
	
	comments := apiGroup.Group("/comments")
	comments.POST("", handler.CreateComment)
	comments.GET("/review/:reviewID", handler.GetCommentsByReview)
	
	return router, mockService
}

func TestCreateRating(t *testing.T) {
	router, mockService := setupTest()
	
	userID := uuid.New()
	serviceID := uuid.New()
	score := 4
	
	// Create a rating object to return
	now := time.Now()
	rating := &model.Rating{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		Score:     score,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Setup mock
	mockService.On("CreateRating", mock.Anything, userID, serviceID, score).Return(rating, nil)
	
	// Create request body
	reqBody := CreateRatingRequest{
		UserID:    userID.String(),
		ServiceID: serviceID.String(),
		Score:     score,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/ratings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response model.Rating
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, rating.ID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, serviceID, response.ServiceID)
	assert.Equal(t, score, response.Score)
	
	mockService.AssertExpectations(t)
}

func TestGetAverageRating(t *testing.T) {
	router, mockService := setupTest()
	
	serviceID := uuid.New()
	
	// Create an average rating object to return
	averageRating := &model.AverageRating{
		ServiceID:    serviceID,
		AverageScore: 4.5,
		TotalRatings: 10,
	}
	
	// Setup mock
	mockService.On("GetAverageRating", mock.Anything, serviceID).Return(averageRating, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/ratings/service/"+serviceID.String()+"/average", nil)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response model.AverageRating
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, serviceID, response.ServiceID)
	assert.Equal(t, 4.5, response.AverageScore)
	assert.Equal(t, 10, response.TotalRatings)
	
	mockService.AssertExpectations(t)
}

func TestCreateReview(t *testing.T) {
	router, mockService := setupTest()
	
	userID := uuid.New()
	serviceID := uuid.New()
	ratingID := uuid.New()
	title := "Test Review"
	content := "This is a test review"
	
	// Create a review object to return
	now := time.Now()
	review := &model.Review{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		RatingID:  ratingID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Setup mock
	mockService.On("CreateReview", mock.Anything, userID, serviceID, ratingID, title, content).Return(review, nil)
	
	// Create request body
	reqBody := CreateReviewRequest{
		UserID:    userID.String(),
		ServiceID: serviceID.String(),
		RatingID:  ratingID.String(),
		Title:     title,
		Content:   content,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/reviews", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response model.Review
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, review.ID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, serviceID, response.ServiceID)
	assert.Equal(t, ratingID, response.RatingID)
	assert.Equal(t, title, response.Title)
	assert.Equal(t, content, response.Content)
	
	mockService.AssertExpectations(t)
}

func TestGetReviewsByService(t *testing.T) {
	router, mockService := setupTest()
	
	serviceID := uuid.New()
	
	// Create a review list to return
	now := time.Now()
	reviews := []*model.ReviewWithRating{
		{
			Review: model.Review{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				ServiceID: serviceID,
				RatingID:  uuid.New(),
				Title:     "Review 1",
				Content:   "This is review 1",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Score: 4,
		},
		{
			Review: model.Review{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				ServiceID: serviceID,
				RatingID:  uuid.New(),
				Title:     "Review 2",
				Content:   "This is review 2",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Score: 5,
		},
	}
	
	// Setup mock
	mockService.On(
		"GetReviewsByService", 
		mock.Anything, 
		serviceID, 
		mock.MatchedBy(func(p pagination.Params) bool {
			return p.Limit == 10 && p.Offset == 0
		}),
	).Return(reviews, 2, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/reviews/service/"+serviceID.String(), nil)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	
	reviewsData, ok := response["reviews"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, reviewsData, 2)
	
	mockService.AssertExpectations(t)
}

func TestCreateComment(t *testing.T) {
	router, mockService := setupTest()
	
	userID := uuid.New()
	reviewID := uuid.New()
	content := "This is a test comment"
	
	// Create a comment object to return
	now := time.Now()
	comment := &model.Comment{
		ID:        uuid.New(),
		UserID:    userID,
		ReviewID:  reviewID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Setup mock
	mockService.On("CreateComment", mock.Anything, userID, reviewID, content).Return(comment, nil)
	
	// Create request body
	reqBody := CreateCommentRequest{
		UserID:   userID.String(),
		ReviewID: reviewID.String(),
		Content:  content,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/comments", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response model.Comment
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, comment.ID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, reviewID, response.ReviewID)
	assert.Equal(t, content, response.Content)
	
	mockService.AssertExpectations(t)
}

func TestGetCommentsByReview(t *testing.T) {
	router, mockService := setupTest()
	
	reviewID := uuid.New()
	
	// Create a comment list to return
	now := time.Now()
	comments := []*model.Comment{
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ReviewID:  reviewID,
			Content:   "Comment 1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ReviewID:  reviewID,
			Content:   "Comment 2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	
	// Setup mock
	mockService.On(
		"GetCommentsByReview", 
		mock.Anything, 
		reviewID, 
		mock.MatchedBy(func(p pagination.Params) bool {
			return p.Limit == 10 && p.Offset == 0
		}),
	).Return(comments, 2, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/comments/review/"+reviewID.String(), nil)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	
	commentsData, ok := response["comments"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, commentsData, 2)
	
	mockService.AssertExpectations(t)
}

func TestGetRatingByUserAndService_NotFound(t *testing.T) {
	router, mockService := setupTest()
	
	userID := uuid.New()
	serviceID := uuid.New()
	
	// Setup mock to return not found error
	mockService.On("GetRatingByUserAndService", mock.Anything, userID, serviceID).
		Return(nil, errors.New("rating not found"))
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/ratings/user/"+userID.String()+"/service/"+serviceID.String(), nil)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response - should return 500 since we're using a generic error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	mockService.AssertExpectations(t)
}
