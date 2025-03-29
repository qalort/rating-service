-- Create tables for ratings, reviews, and comments

-- Create ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    service_id CHAR(36) NOT NULL,
    score INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT chk_score CHECK (score >= 1 AND score <= 5),
    CONSTRAINT unique_user_service UNIQUE (user_id, service_id)
);

-- Create index for service_id for efficient queries
CREATE INDEX IF NOT EXISTS idx_ratings_service_id ON ratings(service_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);

-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    service_id CHAR(36) NOT NULL,
    rating_id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT unique_rating UNIQUE (rating_id),
    FOREIGN KEY (rating_id) REFERENCES ratings(id) ON DELETE CASCADE
);

-- Create indexes for reviews table
CREATE INDEX IF NOT EXISTS idx_reviews_service_id ON reviews(service_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating_id ON reviews(rating_id);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    review_id CHAR(36) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE
);

-- Create indexes for comments table
CREATE INDEX IF NOT EXISTS idx_comments_review_id ON comments(review_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
