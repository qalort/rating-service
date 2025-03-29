# Rating and Review System API

A comprehensive RESTful API for ratings, reviews, and comments, built with Go and PostgreSQL.

## Features

- **User authentication** with JWT
- **Ratings** - Create and retrieve ratings
- **Reviews** - Create detailed reviews with title and content
- **Comments** - Comment on reviews
- **Pagination** - All listing endpoints support pagination
- **Sorting** - Flexible sorting options
- **Swagger documentation** - API fully documented
- **Docker support** - Easy setup with Docker Compose

## Architecture

This project follows hexagonal (ports and adapters) architecture:

- **Domain layer**: Core business logic, models and interfaces
- **Service layer**: Implementation of business logic
- **Infrastructure layer**: External dependencies, database access, HTTP handlers
- **Presentation layer**: API endpoints and response formatting

## API Endpoints

| Method | Endpoint                             | Description                                   | Auth Required |
|--------|--------------------------------------|-----------------------------------------------|--------------|
| POST   | /api/v1/auth/register                | Register a new user                           | No           |
| POST   | /api/v1/auth/login                   | Login a user                                  | No           |
| POST   | /api/v1/ratings                      | Create a new rating                           | Yes          |
| GET    | /api/v1/ratings/service/{serviceID}  | Get all ratings for a service                 | No           |
| GET    | /api/v1/ratings/service/{serviceID}/average | Get average rating for a service       | No           |
| GET    | /api/v1/ratings/service/{serviceID}/me | Get user's rating for a service            | Yes          |
| POST   | /api/v1/reviews                      | Create a new review                           | Yes          |
| GET    | /api/v1/reviews/{reviewID}           | Get a review by ID                            | No           |
| GET    | /api/v1/reviews/service/{serviceID}  | Get all reviews for a service                 | No           |
| POST   | /api/v1/comments                     | Create a new comment                          | Yes          |
| GET    | /api/v1/comments/review/{reviewID}   | Get all comments for a review                 | No           |

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.22 or later (for local development)

### Running with Docker

1. Clone the repository
```bash
git clone https://github.com/yourusername/rating-system.git
cd rating-system
```

2. Start the application with Docker Compose
```bash
docker-compose up -d
```

3. Access the API at http://localhost:8000
4. Access Swagger UI at http://localhost:8080

### Environment Variables

The following environment variables can be configured:

| Variable        | Description                     | Default               |
|-----------------|---------------------------------|-----------------------|
| DB_HOST         | PostgreSQL host                 | postgres              |
| DB_PORT         | PostgreSQL port                 | 5432                  |
| DB_USER         | PostgreSQL username             | postgres              |
| DB_PASSWORD     | PostgreSQL password             | postgres              |
| DB_NAME         | PostgreSQL database name        | ratings               |
| DB_SSLMODE      | PostgreSQL SSL mode             | disable               |
| JWT_SECRET      | Secret key for JWT tokens       | your_jwt_secret_key_change_in_production |
| PORT            | API server port                 | 8000                  |
| GIN_MODE        | Gin mode (debug or release)     | release               |

## Development

### Project Structure

```
.
├── cmd                  # Application entry points
├── docs                 # Swagger documentation
├── internal             # Non-exported code
│   ├── domain           # Business domain
│   │   ├── model        # Domain models
│   │   └── port         # Interfaces (ports)
│   ├── infrastructure   # External facing adapters
│   │   ├── handler      # HTTP handlers
│   │   ├── repository   # Database implementations
│   │   └── auth         # Authentication
│   └── service          # Business logic implementations
├── pkg                  # Exported libraries
├── scripts              # Helper scripts
└── test                 # Test utilities and fixtures
```

### Running Tests

Run all tests:
```bash
go test ./...
```

Run specific tests:
```bash
go test ./internal/infrastructure/handler -v
```

### API Documentation

The API is documented using Swagger (OpenAPI). You can access the documentation at:

- `/swagger/index.html` when running the API
- or http://localhost:8080 when using Docker Compose

## License

This project is licensed under the MIT License - see the LICENSE file for details.