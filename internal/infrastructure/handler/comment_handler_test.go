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

func TestCreateComment(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/comments", func(c *gin.Context) {
		// Simulate authentication middleware
		userID := uuid.New()
		c.Set("userID", userID)
		handler.CreateComment(c)
	})

	// Create test data
	userID := uuid.New()
	reviewID := uuid.New()
	content := "This is a great review, I agree with your assessment."
	comment := model.Comment{
		ID:       uuid.New(),
		UserID:   userID,
		ReviewID: reviewID,
		Content:  content,
	}

	// Setup expectations
	mockService.EXPECT().
		CreateComment(gomock.Any(), gomock.Any(), reviewID, content).
		Return(comment, nil).
		Times(1)

	// Test request
	reqBody, _ := json.Marshal(map[string]interface{}{
		"review_id": reviewID.String(),
		"content":   content,
	})
	req, _ := http.NewRequest("POST", "/comments", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusCreated, resp.Code)
	var respBody model.Comment
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, comment.ID, respBody.ID)
	assert.Equal(t, comment.Content, respBody.Content)
}

func TestGetCommentsByReview(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := logrus.New()
	handler := NewHandler(mockService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/comments/review/:reviewID", handler.GetCommentsByReview)

	// Create test data
	reviewID := uuid.New()
	comments := []model.Comment{
		{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			ReviewID: reviewID,
			Content:  "I agree with this review.",
		},
		{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			ReviewID: reviewID,
			Content:  "Great points in this review.",
		},
	}
	total := 2
	params := pagination.NewParamsWithOffset(10, 0, "created_at", "desc")

	// Setup expectations
	mockService.EXPECT().
		GetCommentsByReview(gomock.Any(), reviewID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, p pagination.Params) ([]model.Comment, int, error) {
			assert.Equal(t, params.GetLimit(), p.GetLimit())
			assert.Equal(t, params.GetOffset(), p.GetOffset())
			return comments, total, nil
		}).
		Times(1)

	// Test request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/comments/review/%s?limit=10&offset=0&sort_by=created_at&sort_direction=desc", reviewID.String()), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["comments"], 2)
	assert.Equal(t, float64(total), respBody["total"])
	assert.Equal(t, float64(10), respBody["limit"])
	assert.Equal(t, float64(0), respBody["offset"])
}