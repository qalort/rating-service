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

func TestMySQLRepository_CreateRating(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Create a test rating
        now := time.Now()
        rating := &model.Rating{
                ID:        uuid.New(),
                UserID:    uuid.New(),
                ServiceID: uuid.New(),
                Score:     5,
                CreatedAt: now,
                UpdatedAt: now,
        }

        // Set up expectations
        mock.ExpectExec("INSERT INTO ratings").
                WithArgs(
                        rating.ID.String(),
                        rating.UserID.String(),
                        rating.ServiceID.String(),
                        rating.Score,
                        rating.CreatedAt,
                        rating.UpdatedAt,
                ).
                WillReturnResult(sqlmock.NewResult(1, 1))

        // Call the function being tested
        err = repo.CreateRating(context.Background(), rating)

        // Assertions
        assert.NoError(t, err)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_GetRatingByID(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Test data
        ratingID := uuid.New()
        userID := uuid.New()
        serviceID := uuid.New()
        score := 5
        createdAt := time.Now()
        updatedAt := time.Now()

        // Set up expectations
        rows := sqlmock.NewRows([]string{"id", "user_id", "service_id", "score", "created_at", "updated_at"}).
                AddRow(ratingID.String(), userID.String(), serviceID.String(), score, createdAt, updatedAt)

        mock.ExpectQuery("SELECT id, user_id, service_id, score, created_at, updated_at FROM ratings WHERE id = ?").
                WithArgs(ratingID.String()).
                WillReturnRows(rows)

        // Call the function being tested
        rating, err := repo.GetRatingByID(context.Background(), ratingID)

        // Assertions
        assert.NoError(t, err)
        assert.Equal(t, ratingID, rating.ID)
        assert.Equal(t, userID, rating.UserID)
        assert.Equal(t, serviceID, rating.ServiceID)
        assert.Equal(t, score, rating.Score)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_GetAverageRatingByService(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Test data
        serviceID := uuid.New()
        averageScore := 4.5
        totalRatings := 10

        // Set up expectations
        rows := sqlmock.NewRows([]string{"average_score", "total_ratings"}).
                AddRow(averageScore, totalRatings)

        mock.ExpectQuery("SELECT AVG\\(score\\) as average_score, COUNT\\(\\*\\) as total_ratings FROM ratings WHERE service_id = ?").
                WithArgs(serviceID.String()).
                WillReturnRows(rows)

        // Call the function being tested
        result, err := repo.GetAverageRatingByService(context.Background(), serviceID)

        // Assertions
        assert.NoError(t, err)
        assert.Equal(t, serviceID, result.ServiceID)
        assert.Equal(t, averageScore, result.AverageScore)
        assert.Equal(t, totalRatings, result.TotalRatings)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_CreateReview(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Create a test review
        now := time.Now()
        review := &model.Review{
                ID:        uuid.New(),
                UserID:    uuid.New(),
                ServiceID: uuid.New(),
                RatingID:  uuid.New(),
                Title:     "Test Review",
                Content:   "This is a test review content.",
                CreatedAt: now,
                UpdatedAt: now,
        }

        // Set up expectations
        mock.ExpectExec("INSERT INTO reviews").
                WithArgs(
                        review.ID.String(),
                        review.UserID.String(),
                        review.ServiceID.String(),
                        review.RatingID.String(),
                        review.Title,
                        review.Content,
                        review.CreatedAt,
                        review.UpdatedAt,
                ).
                WillReturnResult(sqlmock.NewResult(1, 1))

        // Call the function being tested
        err = repo.CreateReview(context.Background(), review)

        // Assertions
        assert.NoError(t, err)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_GetReviewsByService(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Test data
        serviceID := uuid.New()
        params := pagination.NewParams(1, 10, "", "")
        total := 2

        review1ID := uuid.New()
        review1UserID := uuid.New()
        review1RatingID := uuid.New()
        review1Title := "Review 1"
        review1Content := "Content 1"
        review1CreatedAt := time.Now()
        review1UpdatedAt := time.Now()
        review1Score := 5

        review2ID := uuid.New()
        review2UserID := uuid.New()
        review2RatingID := uuid.New()
        review2Title := "Review 2"
        review2Content := "Content 2"
        review2CreatedAt := time.Now()
        review2UpdatedAt := time.Now()
        review2Score := 4

        // Set up expectations for count query
        countRows := sqlmock.NewRows([]string{"count"}).AddRow(total)
        mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM reviews WHERE service_id = ?").
                WithArgs(serviceID.String()).
                WillReturnRows(countRows)

        // Set up expectations for the reviews query
        reviewRows := sqlmock.NewRows([]string{
                "id", "user_id", "service_id", "rating_id", "title", "content", "created_at", "updated_at", "score",
        }).
                AddRow(
                        review1ID.String(),
                        review1UserID.String(),
                        serviceID.String(),
                        review1RatingID.String(),
                        review1Title,
                        review1Content,
                        review1CreatedAt,
                        review1UpdatedAt,
                        review1Score,
                ).
                AddRow(
                        review2ID.String(),
                        review2UserID.String(),
                        serviceID.String(),
                        review2RatingID.String(),
                        review2Title,
                        review2Content,
                        review2CreatedAt,
                        review2UpdatedAt,
                        review2Score,
                )

        mock.ExpectQuery("SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score FROM reviews r JOIN ratings rt ON r.rating_id = rt.id WHERE r.service_id = ? ORDER BY r.created_at DESC LIMIT ? OFFSET ?").
                WithArgs(serviceID.String(), params.GetLimit(), params.GetOffset()).
                WillReturnRows(reviewRows)

        // Call the function being tested
        reviews, count, err := repo.GetReviewsByService(context.Background(), serviceID, params)

        // Assertions
        assert.NoError(t, err)
        assert.Equal(t, total, count)
        assert.Len(t, reviews, 2)
        assert.Equal(t, review1ID, reviews[0].ID)
        assert.Equal(t, review1Title, reviews[0].Title)
        assert.Equal(t, review1Score, reviews[0].Score)
        assert.Equal(t, review2ID, reviews[1].ID)
        assert.Equal(t, review2Title, reviews[1].Title)
        assert.Equal(t, review2Score, reviews[1].Score)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_CreateComment(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Create a test comment
        now := time.Now()
        comment := &model.Comment{
                ID:        uuid.New(),
                UserID:    uuid.New(),
                ReviewID:  uuid.New(),
                Content:   "This is a test comment.",
                CreatedAt: now,
                UpdatedAt: now,
        }

        // Set up expectations
        mock.ExpectExec("INSERT INTO comments").
                WithArgs(
                        comment.ID.String(),
                        comment.UserID.String(),
                        comment.ReviewID.String(),
                        comment.Content,
                        comment.CreatedAt,
                        comment.UpdatedAt,
                ).
                WillReturnResult(sqlmock.NewResult(1, 1))

        // Call the function being tested
        err = repo.CreateComment(context.Background(), comment)

        // Assertions
        assert.NoError(t, err)
        assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMySQLRepository_GetCommentsByReview(t *testing.T) {
        // Create a new mock database connection
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("Failed to create mock database connection: %v", err)
        }
        defer db.Close()

        // Create a test logger
        logger := logrus.New()
        logger.SetLevel(logrus.ErrorLevel)

        // Create a new repository with the mock database
        repo := NewMySQLRepository(db, logger)

        // Test data
        reviewID := uuid.New()
        params := pagination.NewParams(1, 10, "", "")
        total := 2

        comment1ID := uuid.New()
        comment1UserID := uuid.New()
        comment1Content := "Comment 1"
        comment1CreatedAt := time.Now()
        comment1UpdatedAt := time.Now()

        comment2ID := uuid.New()
        comment2UserID := uuid.New()
        comment2Content := "Comment 2"
        comment2CreatedAt := time.Now()
        comment2UpdatedAt := time.Now()

        // Set up expectations for count query
        countRows := sqlmock.NewRows([]string{"count"}).AddRow(total)
        mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM comments WHERE review_id = ?").
                WithArgs(reviewID.String()).
                WillReturnRows(countRows)

        // Set up expectations for the comments query
        commentRows := sqlmock.NewRows([]string{
                "id", "user_id", "review_id", "content", "created_at", "updated_at",
        }).
                AddRow(
                        comment1ID.String(),
                        comment1UserID.String(),
                        reviewID.String(),
                        comment1Content,
                        comment1CreatedAt,
                        comment1UpdatedAt,
                ).
                AddRow(
                        comment2ID.String(),
                        comment2UserID.String(),
                        reviewID.String(),
                        comment2Content,
                        comment2CreatedAt,
                        comment2UpdatedAt,
                )

        mock.ExpectQuery("SELECT id, user_id, review_id, content, created_at, updated_at FROM comments WHERE review_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?").
                WithArgs(reviewID.String(), params.GetLimit(), params.GetOffset()).
                WillReturnRows(commentRows)

        // Call the function being tested
        comments, count, err := repo.GetCommentsByReview(context.Background(), reviewID, params)

        // Assertions
        assert.NoError(t, err)
        assert.Equal(t, total, count)
        assert.Len(t, comments, 2)
        assert.Equal(t, comment1ID, comments[0].ID)
        assert.Equal(t, comment1Content, comments[0].Content)
        assert.Equal(t, comment2ID, comments[1].ID)
        assert.Equal(t, comment2Content, comments[1].Content)
        assert.NoError(t, mock.ExpectationsWereMet())
}