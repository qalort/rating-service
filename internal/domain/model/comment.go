package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Comment represents a user comment on a review
type Comment struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ReviewID  uuid.UUID `json:"review_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewComment creates a new comment with validation
func NewComment(userID, reviewID uuid.UUID, content string) (*Comment, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	if reviewID == uuid.Nil {
		return nil, errors.New("review ID cannot be empty")
	}

	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	now := time.Now()
	return &Comment{
		ID:        uuid.New(),
		UserID:    userID,
		ReviewID:  reviewID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateContent updates the comment content
func (c *Comment) UpdateContent(content string) error {
	if content == "" {
		return errors.New("content cannot be empty")
	}

	c.Content = content
	c.UpdatedAt = time.Now()
	return nil
}
