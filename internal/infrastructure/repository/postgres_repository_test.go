package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"rating-system/internal/domain/model"
	"rating-system/pkg/pagination"
)

func setupMock(t *testing.T) (*PostgresRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	repo := &PostgresRepository{
		db:  db,
		log: logger,
	}

	return repo, mock
}

func TestCreateRating(t *testing.T) {
	repo, mock := setupMock(t)
	ctx := context.Background()

	rating, _ := model.NewRating(uuid.New(), uuid.New(), 5)

	mock.ExpectExec("INSERT INTO ratings").
		WithArgs(rating.ID, rating.UserID, rating.ServiceID, rating.Score, rating.CreatedAt, rating.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateRating(ctx, rating)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetRatingByID(t *testing.T) {
	repo, mock := setupMock(t)
	ctx := context.Background()

	ratingID := uuid.New()
	userID := uuid.New()
	serviceID := uuid.New()
	score := 4
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "service_id", "score", "created_at", "updated_at"}).
		AddRow(ratingID, userID, serviceID, score, now, now)

	mock.ExpectQuery("SELECT (.+) FROM ratings WHERE id = (.+)").
		WithArgs(ratingID).
		WillReturnRows(rows)

	rating, err := repo.GetRatingByID(ctx, ratingID)
	assert.NoError(t, err)
	assert.Equal(t, ratingID, rating.ID)
	assert.Equal(t, userID, rating.UserID)
	assert.Equal(t, serviceID, rating.ServiceID)
	assert.Equal(t, score, rating.Score)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCalculateAverageRating(t *testing.T) {
	repo, mock := setupMock(t)
	ctx := context.Background()

	serviceID := uuid.New()
	avgScore := 4.5
	totalRatings := 10

	rows := sqlmock.NewRows([]string{"average_score", "total_ratings"}).
		AddRow(avgScore, totalRatings)

	mock.ExpectQuery("SELECT AVG\\(score\\) AS average_score, COUNT\\(\\*\\) AS total_ratings FROM ratings WHERE service_id = (.+)").
		WithArgs(serviceID).
		WillReturnRows(rows)

	avgRating, err := repo.CalculateAverageRating(ctx, serviceID)
	assert.NoError(t, err)
	assert.Equal(t, serviceID, avgRating.ServiceID)
	assert.Equal(t, avgScore, avgRating.AverageScore)
	assert.Equal(t, totalRatings, avgRating.TotalRatings)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetReviewsByService(t *testing.T) {
	repo, mock := setupMock(t)
	ctx := context.Background()

	serviceID := uuid.New()
	reviewID := uuid.New()
	userID := uuid.New()
	ratingID := uuid.New()
	title := "Test Review"
	content := "This is a test review"
	score := 4
	now := time.Now()

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM reviews WHERE service_id = (.+)").
		WithArgs(serviceID).
		WillReturnRows(countRows)

	// Mock data query
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "service_id", "rating_id", "title", "content", "created_at", "updated_at", "score",
	}).AddRow(
		reviewID, userID, serviceID, ratingID, title, content, now, now, score,
	)

	mock.ExpectQuery("SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score FROM reviews r").
		WithArgs(serviceID, 10, 0).
		WillReturnRows(rows)

	params := pagination.Params{
		Limit:  10,
		Offset: 0,
	}

	reviews, total, err := repo.GetReviewsByService(ctx, serviceID, params)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, reviews, 1)
	assert.Equal(t, reviewID, reviews[0].ID)
	assert.Equal(t, title, reviews[0].Title)
	assert.Equal(t, score, reviews[0].Score)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateComment(t *testing.T) {
	repo, mock := setupMock(t)
	ctx := context.Background()

	comment, _ := model.NewComment(uuid.New(), uuid.New(), "Test comment")

	mock.ExpectExec("INSERT INTO comments").
		WithArgs(comment.ID, comment.UserID, comment.ReviewID, comment.Content, comment.CreatedAt, comment.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateComment(ctx, comment)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
