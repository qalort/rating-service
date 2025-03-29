package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Rating represents a user rating for a specific service
type Rating struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ServiceID uuid.UUID `json:"service_id"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRating creates a new rating with validation
func NewRating(userID, serviceID uuid.UUID, score int) (*Rating, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	if serviceID == uuid.Nil {
		return nil, errors.New("service ID cannot be empty")
	}

	if score < 1 || score > 5 {
		return nil, errors.New("score must be between 1 and 5")
	}

	now := time.Now()
	return &Rating{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		Score:     score,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateScore updates the rating score with validation
func (r *Rating) UpdateScore(score int) error {
	if score < 1 || score > 5 {
		return errors.New("score must be between 1 and 5")
	}

	r.Score = score
	r.UpdatedAt = time.Now()
	return nil
}

// AverageRating represents the average rating for a service
type AverageRating struct {
	ServiceID   uuid.UUID `json:"service_id"`
	AverageScore float64   `json:"average_score"`
	TotalRatings int       `json:"total_ratings"`
}
