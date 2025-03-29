package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"rating-system/internal/domain/model"
	"rating-system/internal/domain/port/mocks"
	"rating-system/pkg/pagination"
)

func TestCreateReview(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/reviews", func(c *gin.Context) {
		// Simulate authentication middleware
		userID := uuid.New()
		c.Set("userID", userID)
		handler.CreateReview(c)
	})

	// Create test data
	userID := uuid.New()
	serviceID := uuid.New()
	ratingID := uuid.New()
	title := "Great service"
	content := "I was really impressed with the quality of service provided."
	review := model.Review{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		RatingID:  ratingID,
		Title:     title,
		Content:   content,
	}

	// Setup expectations
	mockService.EXPECT().
		CreateReview(gomock.Any(), gomock.Any(), serviceID, ratingID, title, content).
		Return(review, nil).
		Times(1)

	// Test request
	reqBody, _ := json.Marshal(map[string]interface{}{
		"service_id": serviceID.String(),
		"rating_id":  ratingID.String(),
		"title":      title,
		"content":    content,
	})
	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusCreated, resp.Code)
	var respBody model.Review
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, review.ID, respBody.ID)
	assert.Equal(t, review.Title, respBody.Title)
	assert.Equal(t, review.Content, respBody.Content)
}

func TestGetReviewByID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/reviews/:reviewID", handler.GetReviewByID)

	// Create test data
	reviewID := uuid.New()
	review := model.Review{
		ID:        reviewID,
		UserID:    uuid.New(),
		ServiceID: uuid.New(),
		RatingID:  uuid.New(),
		Title:     "Great service",
		Content:   "I was really impressed with the quality of service provided.",
	}

	// Setup expectations
	mockService.EXPECT().
		GetReviewByID(gomock.Any(), reviewID).
		Return(review, nil).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/reviews/%s", reviewID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody model.Review
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, review.ID, respBody.ID)
	assert.Equal(t, review.Title, respBody.Title)
	assert.Equal(t, review.Content, respBody.Content)
}

func TestGetReviewsByService(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/reviews/service/:serviceID", handler.GetReviewsByService)

	// Create test data
	serviceID := uuid.New()
	reviews := []model.Review{
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ServiceID: serviceID,
			RatingID:  uuid.New(),
			Title:     "Great service",
			Content:   "I was really impressed with the quality of service provided.",
		},
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ServiceID: serviceID,
			RatingID:  uuid.New(),
			Title:     "Good experience",
			Content:   "I had a good experience using this service.",
		},
	}
	total := 2
	params := pagination.NewParamsWithOffset(10, 0, "created_at", "desc")

	// Setup expectations
	mockService.EXPECT().
		GetReviewsByService(gomock.Any(), serviceID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, p pagination.Params) ([]model.Review, int, error) {
			assert.Equal(t, params.GetLimit(), p.GetLimit())
			assert.Equal(t, params.GetOffset(), p.GetOffset())
			return reviews, total, nil
		}).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/reviews/service/%s?limit=10&offset=0&sort_by=created_at&sort_direction=desc", serviceID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["reviews"], 2)
	assert.Equal(t, float64(total), respBody["total"])
	assert.Equal(t, float64(10), respBody["limit"])
	assert.Equal(t, float64(0), respBody["offset"])
}