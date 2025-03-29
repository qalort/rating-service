package repository

import (
        "context"
        "database/sql"
        "errors"
        "fmt"
        "strings"

        "github.com/google/uuid"
        "github.com/lib/pq"
        "github.com/sirupsen/logrus"

        "rating-system/internal/domain/model"
        "rating-system/internal/domain/port"
        "rating-system/pkg/pagination"
)

// PostgresRepository implements the Repository port with PostgreSQL
type PostgresRepository struct {
        db  *sql.DB
        log *logrus.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB, log *logrus.Logger) port.Repository {
        return &PostgresRepository{
                db:  db,
                log: log,
        }
}

// execWithContext executes a query with context and logs errors
func (r *PostgresRepository) execWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
        r.log.WithFields(logrus.Fields{
                "query": query,
                "args":  args,
        }).Debug("Executing query")

        result, err := r.db.ExecContext(ctx, query, args...)
        if err != nil {
                r.log.WithError(err).WithFields(logrus.Fields{
                        "query": query,
                        "args":  args,
                }).Error("Query execution failed")
                return nil, err
        }
        return result, nil
}

// queryRowWithContext queries a single row with context and logs errors
func (r *PostgresRepository) queryRowWithContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
        r.log.WithFields(logrus.Fields{
                "query": query,
                "args":  args,
        }).Debug("Querying single row")

        return r.db.QueryRowContext(ctx, query, args...)
}

// queryWithContext queries multiple rows with context and logs errors
func (r *PostgresRepository) queryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
        r.log.WithFields(logrus.Fields{
                "query": query,
                "args":  args,
        }).Debug("Querying multiple rows")

        rows, err := r.db.QueryContext(ctx, query, args...)
        if err != nil {
                r.log.WithError(err).WithFields(logrus.Fields{
                        "query": query,
                        "args":  args,
                }).Error("Query execution failed")
                return nil, err
        }
        return rows, nil
}

// CreateRating creates a new rating in the database
func (r *PostgresRepository) CreateRating(ctx context.Context, rating *model.Rating) error {
        query := `
                INSERT INTO ratings (id, user_id, service_id, score, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
        `
        _, err := r.execWithContext(
                ctx,
                query,
                rating.ID,
                rating.UserID,
                rating.ServiceID,
                rating.Score,
                rating.CreatedAt,
                rating.UpdatedAt,
        )
        if err != nil {
                // Check for unique constraint violation
                if pqErr, ok := err.(*pq.Error); ok {
                        if pqErr.Code == "23505" { // unique_violation
                                return errors.New("rating already exists for this user and service")
                        }
                }
                return err
        }
        return nil
}

// GetRatingByID retrieves a rating by ID
func (r *PostgresRepository) GetRatingByID(ctx context.Context, id uuid.UUID) (*model.Rating, error) {
        query := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE id = $1
        `
        row := r.queryRowWithContext(ctx, query, id)

        var rating model.Rating
        err := row.Scan(
                &rating.ID,
                &rating.UserID,
                &rating.ServiceID,
                &rating.Score,
                &rating.CreatedAt,
                &rating.UpdatedAt,
        )
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("rating not found")
                }
                return nil, err
        }
        return &rating, nil
}

// GetRatingByUserAndService retrieves a rating by user and service
func (r *PostgresRepository) GetRatingByUserAndService(ctx context.Context, userID, serviceID uuid.UUID) (*model.Rating, error) {
        query := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE user_id = $1 AND service_id = $2
        `
        row := r.queryRowWithContext(ctx, query, userID, serviceID)

        var rating model.Rating
        err := row.Scan(
                &rating.ID,
                &rating.UserID,
                &rating.ServiceID,
                &rating.Score,
                &rating.CreatedAt,
                &rating.UpdatedAt,
        )
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("rating not found")
                }
                return nil, err
        }
        return &rating, nil
}

// GetRatingsByService retrieves ratings by service ID with pagination
func (r *PostgresRepository) GetRatingsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.Rating, int, error) {
        // Get total count
        countQuery := `SELECT COUNT(*) FROM ratings WHERE service_id = $1`
        var total int
        err := r.queryRowWithContext(ctx, countQuery, serviceID).Scan(&total)
        if err != nil {
                return nil, 0, err
        }

        // Build the query with sorting and pagination
        baseQuery := `
                SELECT id, user_id, service_id, score, created_at, updated_at
                FROM ratings
                WHERE service_id = $1
        `

        // Add sorting
        sortBy := params.GetSortBy()
        if sortBy != "" {
                baseQuery += fmt.Sprintf(" ORDER BY %s", sanitizeSortField(sortBy))
                if params.GetSortDirection() == "desc" {
                        baseQuery += " DESC"
                } else {
                        baseQuery += " ASC"
                }
        } else {
                // Default sort
                baseQuery += " ORDER BY created_at DESC"
        }

        // Add pagination
        baseQuery += " LIMIT $2 OFFSET $3"

        // Execute the query
        rows, err := r.queryWithContext(
                ctx,
                baseQuery,
                serviceID,
                params.GetLimit(),
                params.GetOffset(),
        )
        if err != nil {
                return nil, 0, err
        }
        defer rows.Close()

        var ratings []*model.Rating
        for rows.Next() {
                var rating model.Rating
                err := rows.Scan(
                        &rating.ID,
                        &rating.UserID,
                        &rating.ServiceID,
                        &rating.Score,
                        &rating.CreatedAt,
                        &rating.UpdatedAt,
                )
                if err != nil {
                        return nil, 0, err
                }
                ratings = append(ratings, &rating)
        }

        if err = rows.Err(); err != nil {
                return nil, 0, err
        }

        return ratings, total, nil
}

// UpdateRating updates an existing rating
func (r *PostgresRepository) UpdateRating(ctx context.Context, rating *model.Rating) error {
        query := `
                UPDATE ratings
                SET score = $1, updated_at = $2
                WHERE id = $3
        `
        _, err := r.execWithContext(
                ctx,
                query,
                rating.Score,
                rating.UpdatedAt,
                rating.ID,
        )
        if err != nil {
                return err
        }
        return nil
}

// CalculateAverageRating calculates the average rating for a service
func (r *PostgresRepository) CalculateAverageRating(ctx context.Context, serviceID uuid.UUID) (*model.AverageRating, error) {
        query := `
                SELECT AVG(score) AS average_score, COUNT(*) AS total_ratings
                FROM ratings
                WHERE service_id = $1
        `
        row := r.queryRowWithContext(ctx, query, serviceID)

        var avgRating model.AverageRating
        var avgScore sql.NullFloat64
        var totalRatings int
        
        err := row.Scan(&avgScore, &totalRatings)
        if err != nil {
                return nil, err
        }

        avgRating.ServiceID = serviceID
        avgRating.TotalRatings = totalRatings
        
        if avgScore.Valid {
                avgRating.AverageScore = avgScore.Float64
        } else {
                avgRating.AverageScore = 0
        }

        return &avgRating, nil
}

// CreateReview creates a new review in the database
func (r *PostgresRepository) CreateReview(ctx context.Context, review *model.Review) error {
        query := `
                INSERT INTO reviews (id, user_id, service_id, rating_id, title, content, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `
        _, err := r.execWithContext(
                ctx,
                query,
                review.ID,
                review.UserID,
                review.ServiceID,
                review.RatingID,
                review.Title,
                review.Content,
                review.CreatedAt,
                review.UpdatedAt,
        )
        if err != nil {
                // Check for unique constraint violation
                if pqErr, ok := err.(*pq.Error); ok {
                        if pqErr.Code == "23505" { // unique_violation
                                return errors.New("review already exists for this rating")
                        }
                }
                return err
        }
        return nil
}

// GetReviewByID retrieves a review by ID
func (r *PostgresRepository) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.ReviewWithRating, error) {
        query := `
                SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score
                FROM reviews r
                JOIN ratings rt ON r.rating_id = rt.id
                WHERE r.id = $1
        `
        row := r.queryRowWithContext(ctx, query, id)

        var review model.ReviewWithRating
        err := row.Scan(
                &review.ID,
                &review.UserID,
                &review.ServiceID,
                &review.RatingID,
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
                return nil, err
        }
        return &review, nil
}

// GetReviewsByService retrieves reviews by service ID with pagination
func (r *PostgresRepository) GetReviewsByService(ctx context.Context, serviceID uuid.UUID, params pagination.Params) ([]*model.ReviewWithRating, int, error) {
        // Get total count
        countQuery := `SELECT COUNT(*) FROM reviews WHERE service_id = $1`
        var total int
        err := r.queryRowWithContext(ctx, countQuery, serviceID).Scan(&total)
        if err != nil {
                return nil, 0, err
        }

        // Build the query with sorting and pagination
        baseQuery := `
                SELECT r.id, r.user_id, r.service_id, r.rating_id, r.title, r.content, r.created_at, r.updated_at, rt.score
                FROM reviews r
                JOIN ratings rt ON r.rating_id = rt.id
                WHERE r.service_id = $1
        `

        // Add sorting
        if params.GetSortBy() != "" {
                // Prefix fields with table alias to avoid ambiguity
                sortField := sanitizeSortField(params.GetSortBy())
                if sortField == "score" {
                        sortField = "rt.score"
                } else {
                        sortField = "r." + sortField
                }
                
                baseQuery += fmt.Sprintf(" ORDER BY %s", sortField)
                if params.GetSortDirection() == "desc" {
                        baseQuery += " DESC"
                } else {
                        baseQuery += " ASC"
                }
        } else {
                // Default sort
                baseQuery += " ORDER BY r.created_at DESC"
        }

        // Add pagination
        baseQuery += " LIMIT $2 OFFSET $3"

        // Execute the query
        rows, err := r.queryWithContext(
                ctx,
                baseQuery,
                serviceID,
                params.GetLimit(),
                params.GetOffset(),
        )
        if err != nil {
                return nil, 0, err
        }
        defer rows.Close()

        var reviews []*model.ReviewWithRating
        for rows.Next() {
                var review model.ReviewWithRating
                err := rows.Scan(
                        &review.ID,
                        &review.UserID,
                        &review.ServiceID,
                        &review.RatingID,
                        &review.Title,
                        &review.Content,
                        &review.CreatedAt,
                        &review.UpdatedAt,
                        &review.Score,
                )
                if err != nil {
                        return nil, 0, err
                }
                reviews = append(reviews, &review)
        }

        if err = rows.Err(); err != nil {
                return nil, 0, err
        }

        return reviews, total, nil
}

// UpdateReview updates an existing review
func (r *PostgresRepository) UpdateReview(ctx context.Context, review *model.Review) error {
        query := `
                UPDATE reviews
                SET title = $1, content = $2, updated_at = $3
                WHERE id = $4
        `
        _, err := r.execWithContext(
                ctx,
                query,
                review.Title,
                review.Content,
                review.UpdatedAt,
                review.ID,
        )
        if err != nil {
                return err
        }
        return nil
}

// CreateComment creates a new comment in the database
func (r *PostgresRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
        query := `
                INSERT INTO comments (id, user_id, review_id, content, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
        `
        _, err := r.execWithContext(
                ctx,
                query,
                comment.ID,
                comment.UserID,
                comment.ReviewID,
                comment.Content,
                comment.CreatedAt,
                comment.UpdatedAt,
        )
        if err != nil {
                return err
        }
        return nil
}

// GetCommentByID retrieves a comment by ID
func (r *PostgresRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
        query := `
                SELECT id, user_id, review_id, content, created_at, updated_at
                FROM comments
                WHERE id = $1
        `
        row := r.queryRowWithContext(ctx, query, id)

        var comment model.Comment
        err := row.Scan(
                &comment.ID,
                &comment.UserID,
                &comment.ReviewID,
                &comment.Content,
                &comment.CreatedAt,
                &comment.UpdatedAt,
        )
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("comment not found")
                }
                return nil, err
        }
        return &comment, nil
}

// GetCommentsByReview retrieves comments by review ID with pagination
func (r *PostgresRepository) GetCommentsByReview(ctx context.Context, reviewID uuid.UUID, params pagination.Params) ([]*model.Comment, int, error) {
        // Get total count
        countQuery := `SELECT COUNT(*) FROM comments WHERE review_id = $1`
        var total int
        err := r.queryRowWithContext(ctx, countQuery, reviewID).Scan(&total)
        if err != nil {
                return nil, 0, err
        }

        // Build the query with sorting and pagination
        baseQuery := `
                SELECT id, user_id, review_id, content, created_at, updated_at
                FROM comments
                WHERE review_id = $1
        `

        // Add sorting
        if params.GetSortBy() != "" {
                baseQuery += fmt.Sprintf(" ORDER BY %s", sanitizeSortField(params.GetSortBy()))
                if params.GetSortDirection() == "desc" {
                        baseQuery += " DESC"
                } else {
                        baseQuery += " ASC"
                }
        } else {
                // Default sort
                baseQuery += " ORDER BY created_at ASC"
        }

        // Add pagination
        baseQuery += " LIMIT $2 OFFSET $3"

        // Execute the query
        rows, err := r.queryWithContext(
                ctx,
                baseQuery,
                reviewID,
                params.GetLimit(),
                params.GetOffset(),
        )
        if err != nil {
                return nil, 0, err
        }
        defer rows.Close()

        var comments []*model.Comment
        for rows.Next() {
                var comment model.Comment
                err := rows.Scan(
                        &comment.ID,
                        &comment.UserID,
                        &comment.ReviewID,
                        &comment.Content,
                        &comment.CreatedAt,
                        &comment.UpdatedAt,
                )
                if err != nil {
                        return nil, 0, err
                }
                comments = append(comments, &comment)
        }

        if err = rows.Err(); err != nil {
                return nil, 0, err
        }

        return comments, total, nil
}

// UpdateComment updates an existing comment
func (r *PostgresRepository) UpdateComment(ctx context.Context, comment *model.Comment) error {
        query := `
                UPDATE comments
                SET content = $1, updated_at = $2
                WHERE id = $3
        `
        _, err := r.execWithContext(
                ctx,
                query,
                comment.Content,
                comment.UpdatedAt,
                comment.ID,
        )
        if err != nil {
                return err
        }
        return nil
}

// sanitizeSortField ensures sort field is safe and exists in the database
func sanitizeSortField(field string) string {
        // List of allowed sort fields
        allowedFields := map[string]bool{
                "score":      true,
                "created_at": true,
                "updated_at": true,
                "title":      true,
                "content":    true,
        }

        // Convert to lowercase and check if allowed
        field = strings.ToLower(field)
        if allowedFields[field] {
                return field
        }

        // Default to created_at if field is not allowed
        return "created_at"
}
