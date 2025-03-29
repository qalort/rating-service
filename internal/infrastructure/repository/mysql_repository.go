package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"rating-system/internal/domain/model"
	"rating-system/internal/domain/port"
	"rating-system/pkg/pagination"
)

// MySQLRepository implements the Repository port with MySQL
type MySQLRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(db *sql.DB, log *logrus.Logger) port.Repository {
	return &MySQLRepository{
		db:     db,
		logger: log,
	}
}

func (r *MySQLRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, username, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.execWithContext(ctx, query,
		user.ID.String(),
		user.Email,
		user.Username,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("failed to create user")
	}

	return nil
}

// execWithContext executes a query with context and logs errors
func (r *MySQLRepository) execWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	r.logger.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("Executing query")

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"query": query,
			"args":  args,
		}).Error("Query execution failed")
		return nil, err
	}
	return result, nil
}

// CreateRating creates a new rating
func (r *MySQLRepository) CreateRating(ctx context.Context, rating *model.Rating) error {
	query := `
                INSERT INTO ratings (id, user_id, service_id, score, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?)
        `

	_, err := r.execWithContext(ctx, query,
		rating.ID.String(),
		rating.UserID.String(),
		rating.ServiceID.String(),
		rating.Score,
		rating.CreatedAt,
		rating.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate entry error (unique constraint violation)
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "unique_user_service") {
			return errors.New("user has already rated this service")
		}
		return fmt.Errorf("failed to create rating: %w", err)
	}

	return nil
}

// UpdateRating updates an existing rating
func (r *MySQLRepository) UpdateRating(ctx context.Context, rating *model.Rating) error {
	query := `
                UPDATE ratings
                SET score = ?, updated_at = ?
                WHERE id = ?
        `

	result, err := r.execWithContext(ctx, query,
		rating.Score,
		rating.UpdatedAt,
		rating.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("rating not found")
	}

	return nil
}

// GetRatingByID retrieves a rating by ID
func (r *MySQLRepository) GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error) {
	query := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE id = ?
        `

	var rating model.Rating
	var idStr, userIDStr, serviceIDStr string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&userIDStr,
		&serviceIDStr,
		&rating.Score,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("rating not found")
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	// Parse UUIDs
	rating.ID, _ = uuid.Parse(idStr)
	rating.UserID, _ = uuid.Parse(userIDStr)
	rating.ServiceID, _ = uuid.Parse(serviceIDStr)

	return &rating, nil
}

// GetRatingsByService retrieves ratings for a specific service with pagination
func (r *MySQLRepository) GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error) {
	// Convert pagination.Params to *pagination.Pagination
	page := pagination.NewPagination(params.GetPage(), params.GetLimit())
	// Count total ratings for this service
	countQuery := `
                SELECT COUNT(*) FROM ratings WHERE service_id = ?
        `
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, serviceID.String()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	// Get paginated ratings
	query := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE service_id = ?
                ORDER BY created_at DESC
                LIMIT ? OFFSET ?
        `

	rows, err := r.db.QueryContext(ctx, query, serviceID.String(), page.GetLimit(), page.GetOffset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ratings: %w", err)
	}
	defer rows.Close()

	var ratings []*model.Rating
	for rows.Next() {
		var rating model.Rating
		var idStr, userIDStr, serviceIDStr string

		if err := rows.Scan(
			&idStr,
			&userIDStr,
			&serviceIDStr,
			&rating.Score,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan rating row: %w", err)
		}

		// Parse UUIDs
		rating.ID, _ = uuid.Parse(idStr)
		rating.UserID, _ = uuid.Parse(userIDStr)
		rating.ServiceID, _ = uuid.Parse(serviceIDStr)

		ratings = append(ratings, &rating)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rating rows: %w", err)
	}

	return ratings, total, nil
}

// CalculateAverageRating calculates the average rating for a service
func (r *MySQLRepository) CalculateAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error) {
	query := `
                SELECT 
                        AVG(score) as average_score, 
                        COUNT(*) as total_ratings
                FROM ratings
                WHERE service_id = ?
        `

	var avg sql.NullFloat64
	var count int

	err := r.db.QueryRowContext(ctx, query, serviceID.String()).Scan(&avg, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to get average rating: %w", err)
	}

	// If no ratings yet, avg will be NULL
	averageScore := 0.0
	if avg.Valid {
		averageScore = avg.Float64
	}

	return &model.AverageRating{
		ServiceID:    serviceID,
		AverageScore: averageScore,
		TotalRatings: count,
	}, nil
}

// GetRatingByUserAndService retrieves a rating for a specific user and service
func (r *MySQLRepository) GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error) {
	query := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE user_id = ? AND service_id = ?
        `

	var rating model.Rating
	var idStr, userIDStr, serviceIDStr string

	err := r.db.QueryRowContext(ctx, query, userID.String(), serviceID.String()).Scan(
		&idStr,
		&userIDStr,
		&serviceIDStr,
		&rating.Score,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("rating not found")
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	// Parse UUIDs
	rating.ID, _ = uuid.Parse(idStr)
	rating.UserID, _ = uuid.Parse(userIDStr)
	rating.ServiceID, _ = uuid.Parse(serviceIDStr)

	return &rating, nil
}

// CreateReview creates a new review
func (r *MySQLRepository) CreateReview(ctx context.Context, review *model.Review) error {
	query := `
                INSERT INTO reviews (id, user_id, service_id, rating_id, title, content, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `

	_, err := r.execWithContext(ctx, query,
		review.ID.String(),
		review.UserID.String(),
		review.ServiceID.String(),
		review.RatingID.String(),
		review.Title,
		review.Content,
		review.CreatedAt,
		review.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "unique_rating") {
			return errors.New("a review already exists for this rating")
		}
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// GetReviewByID retrieves a review by ID
func (r *MySQLRepository) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error) {
	query := `
                SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score
                FROM reviews r
                JOIN ratings rt ON r.rating_id = rt.id
                WHERE r.id = ?
        `

	var review model.ReviewWithRating
	var idStr, userIDStr, serviceIDStr, ratingIDStr string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&userIDStr,
		&serviceIDStr,
		&ratingIDStr,
		&review.Title,
		&review.Content,
		&review.CreatedAt,
		&review.UpdatedAt,
		&review.Score,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	// Parse UUIDs
	review.ID, _ = uuid.Parse(idStr)
	review.UserID, _ = uuid.Parse(userIDStr)
	review.ServiceID, _ = uuid.Parse(serviceIDStr)
	review.RatingID, _ = uuid.Parse(ratingIDStr)

	return &review, nil
}

// GetReviewsByService retrieves reviews for a specific service with pagination
func (r *MySQLRepository) GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error) {
	// Convert pagination.Params to *pagination.Pagination
	page := pagination.NewPagination(params.GetPage(), params.GetLimit())
	// Count total reviews for this service
	countQuery := `
                SELECT COUNT(*) FROM reviews WHERE service_id = ?
        `
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, serviceID.String()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reviews: %w", err)
	}

	// Get paginated reviews
	query := `
                SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score
                FROM reviews r
                JOIN ratings rt ON r.rating_id = rt.id
                WHERE r.service_id = ?
                ORDER BY r.created_at DESC
                LIMIT ? OFFSET ?
        `

	rows, err := r.db.QueryContext(ctx, query, serviceID.String(), page.GetLimit(), page.GetOffset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*model.ReviewWithRating
	for rows.Next() {
		var review model.ReviewWithRating
		var idStr, userIDStr, serviceIDStr, ratingIDStr string

		if err := rows.Scan(
			&idStr,
			&userIDStr,
			&serviceIDStr,
			&ratingIDStr,
			&review.Title,
			&review.Content,
			&review.CreatedAt,
			&review.UpdatedAt,
			&review.Score,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan review row: %w", err)
		}

		// Parse UUIDs
		review.ID, _ = uuid.Parse(idStr)
		review.UserID, _ = uuid.Parse(userIDStr)
		review.ServiceID, _ = uuid.Parse(serviceIDStr)
		review.RatingID, _ = uuid.Parse(ratingIDStr)

		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating review rows: %w", err)
	}

	return reviews, total, nil
}

// CreateComment creates a new comment
func (r *MySQLRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	query := `
                INSERT INTO comments (id, user_id, review_id, content, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?)
        `

	_, err := r.execWithContext(ctx, query,
		comment.ID.String(),
		comment.UserID.String(),
		comment.ReviewID.String(),
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetCommentsByReview retrieves comments for a specific review with pagination
func (r *MySQLRepository) GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error) {
	// Convert pagination.Params to *pagination.Pagination
	page := pagination.NewPagination(params.GetPage(), params.GetLimit())
	// Count total comments for this review
	countQuery := `
                SELECT COUNT(*) FROM comments WHERE review_id = ?
        `
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, reviewID.String()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Get paginated comments
	query := `
                SELECT id, user_id, review_id, content, created_at, updated_at
                FROM comments
                WHERE review_id = ?
                ORDER BY created_at ASC
                LIMIT ? OFFSET ?
        `

	rows, err := r.db.QueryContext(ctx, query, reviewID.String(), page.GetLimit(), page.GetOffset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		var idStr, userIDStr, reviewIDStr string

		if err := rows.Scan(
			&idStr,
			&userIDStr,
			&reviewIDStr,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment row: %w", err)
		}

		// Parse UUIDs
		comment.ID, _ = uuid.Parse(idStr)
		comment.UserID, _ = uuid.Parse(userIDStr)
		comment.ReviewID, _ = uuid.Parse(reviewIDStr)

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating comment rows: %w", err)
	}

	return comments, total, nil
}

// GetCommentByID retrieves a comment by ID
func (r *MySQLRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	query := `
                SELECT id, user_id, review_id, content, created_at, updated_at
                FROM comments
                WHERE id = ?
        `

	var comment model.Comment
	var idStr, userIDStr, reviewIDStr string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&userIDStr,
		&reviewIDStr,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Parse UUIDs
	comment.ID, _ = uuid.Parse(idStr)
	comment.UserID, _ = uuid.Parse(userIDStr)
	comment.ReviewID, _ = uuid.Parse(reviewIDStr)

	return &comment, nil
}

// UpdateComment updates a comment
func (r *MySQLRepository) UpdateComment(ctx context.Context, comment *model.Comment) error {
	query := `
                UPDATE comments
                SET content = ?, updated_at = ?
                WHERE id = ?
        `

	result, err := r.execWithContext(ctx, query,
		comment.Content,
		comment.UpdatedAt,
		comment.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("comment not found")
	}

	return nil
}

// UpdateReview updates a review
func (r *MySQLRepository) UpdateReview(ctx context.Context, review *model.Review) error {
	query := `
                UPDATE reviews
                SET title = ?, content = ?, updated_at = ?
                WHERE id = ?
        `

	result, err := r.execWithContext(ctx, query,
		review.Title,
		review.Content,
		review.UpdatedAt,
		review.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("review not found")
	}

	return nil
}