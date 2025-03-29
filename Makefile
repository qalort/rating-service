.PHONY: build test clean run docker-build docker-up docker-down

# Build the application
build:
	go build -o bin/rating-system .

# Run the application
run:
	go run main.go

# Run all tests
test:
	go test -v ./...

# Run a specific test package
test-pkg:
	@echo "Running tests for package: $(pkg)"
	go test -v ./$(pkg)

# Clean build artifacts
clean:
	rm -rf bin/*

# Generate mocks for testing
mocks:
	mockgen -source=internal/domain/port/service.go -destination=internal/domain/port/mocks/mock_service.go
	mockgen -source=internal/domain/port/auth.go -destination=internal/domain/port/mocks/mock_auth_service.go
	mockgen -source=internal/domain/port/repository.go -destination=internal/domain/port/mocks/mock_repository.go

# Build the Docker image
docker-build:
	docker-compose build

# Start the application with Docker Compose
docker-up:
	docker-compose up -d

# Stop and remove Docker Compose services
docker-down:
	docker-compose down

# Show logs from Docker Compose services
docker-logs:
	docker-compose logs -f

# Show help message
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run all tests"
	@echo "  test-pkg      - Run tests for a specific package (e.g. make test-pkg pkg=internal/service)"
	@echo "  clean         - Clean build artifacts"
	@echo "  mocks         - Generate mock implementations for testing"
	@echo "  docker-build  - Build the Docker image"
	@echo "  docker-up     - Start the application with Docker Compose"
	@echo "  docker-down   - Stop and remove Docker Compose services"
	@echo "  docker-logs   - Show logs from Docker Compose services"
	@echo "  help          - Show this help message"