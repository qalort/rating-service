-- Create tables for ratings, reviews, and comments

-- Create ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    service_id UUID NOT NULL,
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT unique_user_service UNIQUE (user_id, service_id)
);

-- Create index for service_id for efficient queries
CREATE INDEX IF NOT EXISTS idx_ratings_service_id ON ratings(service_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);

-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    service_id UUID NOT NULL,
    rating_id UUID NOT NULL REFERENCES ratings(id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT unique_rating UNIQUE (rating_id)
);

-- Create indexes for reviews table
CREATE INDEX IF NOT EXISTS idx_reviews_service_id ON reviews(service_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating_id ON reviews(rating_id);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    review_id UUID NOT NULL REFERENCES reviews(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes for comments table
CREATE INDEX IF NOT EXISTS idx_comments_review_id ON comments(review_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
