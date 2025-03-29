package handler

import (
        "errors"
        "net/http"
        "strconv"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/sirupsen/logrus"

        "rating-system/internal/domain/port"
        "rating-system/pkg/pagination"
        "rating-system/pkg/validator"
)

// Handler handles HTTP requests
type Handler struct {
        service port.Service
        log     *logrus.Logger
}

// NewHandler creates a new handler
func NewHandler(service port.Service, log *logrus.Logger) *Handler {
        return &Handler{
                service: service,
                log:     log,
        }
}

// CreateRatingRequest is the request for creating a rating
type CreateRatingRequest struct {
        ServiceID string `json:"service_id" binding:"required,uuid4"`
        Score     int    `json:"score" binding:"required,min=1,max=5"`
}

// CreateRating handles the creation of a new rating
func (h *Handler) CreateRating(c *gin.Context) {
        var req CreateRatingRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                h.log.WithError(err).Error("Invalid request body")
                c.JSON(http.StatusBadRequest, gin.H{"error": validator.FormatValidationErrors(err)})
                return
        }

        // Get authenticated user ID from context
        userIDVal, exists := c.Get("userID")
        if !exists {
                h.log.Error("User ID not found in context")
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
                return
        }
        
        userID, ok := userIDVal.(uuid.UUID)
        if !ok {
                h.log.Error("Invalid user ID in context")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
                return
        }

        serviceID, err := uuid.Parse(req.ServiceID)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        rating, err := h.service.CreateRating(c.Request.Context(), userID, serviceID, req.Score)
        if err != nil {
                h.log.WithError(err).Error("Failed to create rating")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, rating)
}

// GetRatingsByService handles retrieving ratings for a service
func (h *Handler) GetRatingsByService(c *gin.Context) {
        serviceIDStr := c.Param("serviceID")
        serviceID, err := uuid.Parse(serviceIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        params := extractPaginationParams(c)
        
        ratings, total, err := h.service.GetRatingsByService(c.Request.Context(), serviceID, params)
        if err != nil {
                h.log.WithError(err).Error("Failed to get ratings")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "ratings": ratings,
                "total":   total,
                "limit":   params.GetLimit(),
                "offset":  params.GetOffset(),
        })
}

// GetAverageRating handles retrieving the average rating for a service
func (h *Handler) GetAverageRating(c *gin.Context) {
        serviceIDStr := c.Param("serviceID")
        serviceID, err := uuid.Parse(serviceIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        average, err := h.service.GetAverageRating(c.Request.Context(), serviceID)
        if err != nil {
                h.log.WithError(err).Error("Failed to get average rating")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, average)
}

// GetUserRating handles retrieving a user's rating for a service
func (h *Handler) GetUserRating(c *gin.Context) {
        // Get authenticated user ID from context
        userIDVal, exists := c.Get("userID")
        if !exists {
                h.log.Error("User ID not found in context")
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
                return
        }
        
        userID, ok := userIDVal.(uuid.UUID)
        if !ok {
                h.log.Error("Invalid user ID in context")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
                return
        }

        serviceIDStr := c.Param("serviceID")
        serviceID, err := uuid.Parse(serviceIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        rating, err := h.service.GetRatingByUserAndService(c.Request.Context(), userID, serviceID)
        if err != nil {
                if errors.Is(err, errors.New("rating not found")) {
                        c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
                        return
                }
                h.log.WithError(err).Error("Failed to get user rating")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, rating)
}

// CreateReviewRequest is the request for creating a review
type CreateReviewRequest struct {
        ServiceID string `json:"service_id" binding:"required,uuid4"`
        RatingID  string `json:"rating_id" binding:"required,uuid4"`
        Title     string `json:"title" binding:"required,min=1,max=255"`
        Content   string `json:"content" binding:"required,min=1"`
}

// CreateReview handles the creation of a new review
func (h *Handler) CreateReview(c *gin.Context) {
        var req CreateReviewRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                h.log.WithError(err).Error("Invalid request body")
                c.JSON(http.StatusBadRequest, gin.H{"error": validator.FormatValidationErrors(err)})
                return
        }

        // Get authenticated user ID from context
        userIDVal, exists := c.Get("userID")
        if !exists {
                h.log.Error("User ID not found in context")
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
                return
        }
        
        userID, ok := userIDVal.(uuid.UUID)
        if !ok {
                h.log.Error("Invalid user ID in context")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
                return
        }

        serviceID, err := uuid.Parse(req.ServiceID)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        ratingID, err := uuid.Parse(req.RatingID)
        if err != nil {
                h.log.WithError(err).Error("Invalid rating ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
                return
        }

        review, err := h.service.CreateReview(c.Request.Context(), userID, serviceID, ratingID, req.Title, req.Content)
        if err != nil {
                h.log.WithError(err).Error("Failed to create review")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, review)
}

// GetReviewByID handles retrieving a review by ID
func (h *Handler) GetReviewByID(c *gin.Context) {
        reviewIDStr := c.Param("reviewID")
        reviewID, err := uuid.Parse(reviewIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid review ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
                return
        }

        review, err := h.service.GetReviewByID(c.Request.Context(), reviewID)
        if err != nil {
                if errors.Is(err, errors.New("review not found")) {
                        c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
                        return
                }
                h.log.WithError(err).Error("Failed to get review")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, review)
}

// GetReviewsByService handles retrieving reviews for a service
func (h *Handler) GetReviewsByService(c *gin.Context) {
        serviceIDStr := c.Param("serviceID")
        serviceID, err := uuid.Parse(serviceIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid service ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
                return
        }

        params := extractPaginationParams(c)
        
        reviews, total, err := h.service.GetReviewsByService(c.Request.Context(), serviceID, params)
        if err != nil {
                h.log.WithError(err).Error("Failed to get reviews")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "reviews": reviews,
                "total":   total,
                "limit":   params.GetLimit(),
                "offset":  params.GetOffset(),
        })
}

// CreateCommentRequest is the request for creating a comment
type CreateCommentRequest struct {
        ReviewID string `json:"review_id" binding:"required,uuid4"`
        Content  string `json:"content" binding:"required,min=1"`
}

// CreateComment handles the creation of a new comment
func (h *Handler) CreateComment(c *gin.Context) {
        var req CreateCommentRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                h.log.WithError(err).Error("Invalid request body")
                c.JSON(http.StatusBadRequest, gin.H{"error": validator.FormatValidationErrors(err)})
                return
        }

        // Get authenticated user ID from context
        userIDVal, exists := c.Get("userID")
        if !exists {
                h.log.Error("User ID not found in context")
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
                return
        }
        
        userID, ok := userIDVal.(uuid.UUID)
        if !ok {
                h.log.Error("Invalid user ID in context")
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
                return
        }

        reviewID, err := uuid.Parse(req.ReviewID)
        if err != nil {
                h.log.WithError(err).Error("Invalid review ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
                return
        }

        comment, err := h.service.CreateComment(c.Request.Context(), userID, reviewID, req.Content)
        if err != nil {
                h.log.WithError(err).Error("Failed to create comment")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, comment)
}

// GetCommentsByReview handles retrieving comments for a review
func (h *Handler) GetCommentsByReview(c *gin.Context) {
        reviewIDStr := c.Param("reviewID")
        reviewID, err := uuid.Parse(reviewIDStr)
        if err != nil {
                h.log.WithError(err).Error("Invalid review ID")
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
                return
        }

        params := extractPaginationParams(c)
        
        comments, total, err := h.service.GetCommentsByReview(c.Request.Context(), reviewID, params)
        if err != nil {
                h.log.WithError(err).Error("Failed to get comments")
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "comments": comments,
                "total":    total,
                "limit":    params.GetLimit(),
                "offset":   params.GetOffset(),
        })
}

// extractPaginationParams extracts pagination parameters from the request
func extractPaginationParams(c *gin.Context) pagination.Params {
        limitStr := c.DefaultQuery("limit", "10")
        offsetStr := c.DefaultQuery("offset", "0")
        sortBy := c.DefaultQuery("sort_by", "created_at")
        sortDirection := c.DefaultQuery("sort_direction", "desc")

        limit, err := strconv.Atoi(limitStr)
        if err != nil || limit < 1 {
                limit = 10
        }

        offset, err := strconv.Atoi(offsetStr)
        if err != nil || offset < 0 {
                offset = 0
        }

        return pagination.NewParamsWithOffset(limit, offset, sortBy, sortDirection)
}
