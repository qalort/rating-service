package handler

import (
	"bytes"
	"encoding/json"
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
)

func TestRegister(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	logger := logrus.New()
	handler := NewAuthHandler(mockAuthService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create test data
	username := "testuser"
	email := "test@example.com"
	password := "password123"
	userID := uuid.New()
	token := "jwt-token"

	user := model.User{
		ID:       userID,
		Username: username,
		Email:    email,
	}

	// Setup expectations
	mockAuthService.EXPECT().
		Register(gomock.Any(), username, email, password).
		Return(user, token, nil).
		Times(1)

	// Test request
	reqBody, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
	})
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusCreated, resp.Code)
	var respBody map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, token, respBody["token"])
	
	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, username, userData["username"])
	assert.Equal(t, email, userData["email"])
}

func TestLogin(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	logger := logrus.New()
	handler := NewAuthHandler(mockAuthService, logger)

	// Test case
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/login", handler.Login)

	// Create test data
	email := "test@example.com"
	password := "password123"
	userID := uuid.New()
	token := "jwt-token"
	username := "testuser"

	user := model.User{
		ID:       userID,
		Username: username,
		Email:    email,
	}

	// Setup expectations
	mockAuthService.EXPECT().
		Login(gomock.Any(), email, password).
		Return(user, token, nil).
		Times(1)

	// Test request
	reqBody, _ := json.Marshal(map[string]interface{}{
		"email":    email,
		"password": password,
	})
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verify
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, token, respBody["token"])
	
	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, username, userData["username"])
	assert.Equal(t, email, userData["email"])
}