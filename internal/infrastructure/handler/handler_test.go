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

func TestCreateRating(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/ratings", func(c *gin.Context) {
		// Simulate authentication middleware
		userID := uuid.New()
		c.Set("userID", userID)
		handler.CreateRating(c)
	})

	// Create test data
	serviceID := uuid.New()
	rating := model.Rating{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ServiceID: serviceID,
		Score:     5,
	}

	// Setup expectations
	mockService.EXPECT().
		CreateRating(gomock.Any(), gomock.Any(), serviceID, 5).
		Return(rating, nil).
		Times(1)

	// Test request
	reqBody, _ := json.Marshal(map[string]interface{}{
		"service_id": serviceID.String(),
		"score":      5,
	})
	req, _ := http.NewRequest("POST", "/ratings", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusCreated, resp.Code)
	var respBody model.Rating
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, rating.ID, respBody.ID)
	assert.Equal(t, rating.Score, respBody.Score)
}

func TestGetRatingsByService(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ratings/service/:serviceID", handler.GetRatingsByService)

	// Create test data
	serviceID := uuid.New()
	ratings := []model.Rating{
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ServiceID: serviceID,
			Score:     5,
		},
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ServiceID: serviceID,
			Score:     4,
		},
	}
	total := 2
	params := pagination.NewParamsWithOffset(10, 0, "created_at", "desc")

	// Setup expectations
	mockService.EXPECT().
		GetRatingsByService(gomock.Any(), serviceID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, p pagination.Params) ([]model.Rating, int, error) {
			assert.Equal(t, params.GetLimit(), p.GetLimit())
			assert.Equal(t, params.GetOffset(), p.GetOffset())
			return ratings, total, nil
		}).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/ratings/service/%s?limit=10&offset=0&sort_by=created_at&sort_direction=desc", serviceID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["ratings"], 2)
	assert.Equal(t, float64(total), respBody["total"])
	assert.Equal(t, float64(10), respBody["limit"])
	assert.Equal(t, float64(0), respBody["offset"])
}

func TestGetAverageRating(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ratings/service/:serviceID/average", handler.GetAverageRating)

	// Create test data
	serviceID := uuid.New()
	avgRating := 4.5

	// Setup expectations
	mockService.EXPECT().
		GetAverageRating(gomock.Any(), serviceID).
		Return(avgRating, nil).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/ratings/service/%s/average", serviceID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody float64
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, avgRating, respBody)
}

func TestGetUserRating(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ratings/service/:serviceID/me", func(c *gin.Context) {
		// Simulate authentication middleware
		userID := uuid.New()
		c.Set("userID", userID)
		handler.GetUserRating(c)
	})

	// Create test data
	userID := uuid.New()
	serviceID := uuid.New()
	rating := model.Rating{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		Score:     4,
	}

	// Setup expectations
	mockService.EXPECT().
		GetRatingByUserAndService(gomock.Any(), gomock.Any(), serviceID).
		Return(rating, nil).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/ratings/service/%s/me", serviceID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody model.Rating
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, rating.ID, respBody.ID)
	assert.Equal(t, rating.Score, respBody.Score)
}