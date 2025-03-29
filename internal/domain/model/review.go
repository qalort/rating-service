package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Review represents a user review for a specific service
type Review struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ServiceID uuid.UUID `json:"service_id"`
	RatingID  uuid.UUID `json:"rating_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewReview creates a new review with validation
func NewReview(userID, serviceID, ratingID uuid.UUID, title, content string) (*Review, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	if serviceID == uuid.Nil {
		return nil, errors.New("service ID cannot be empty")
	}

	if ratingID == uuid.Nil {
		return nil, errors.New("rating ID cannot be empty")
	}

	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	now := time.Now()
	return &Review{
		ID:        uuid.New(),
		UserID:    userID,
		ServiceID: serviceID,
		RatingID:  ratingID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateContent updates the review content
func (r *Review) UpdateContent(title, content string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}

	if content == "" {
		return errors.New("content cannot be empty")
	}

	r.Title = title
	r.Content = content
	r.UpdatedAt = time.Now()
	return nil
}

// ReviewWithRating represents a review with its associated rating
type ReviewWithRating struct {
	Review
	Score int `json:"score"`
}
