# Start from the official Go image
FROM golang:1.21-alpine as builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git mysql-client

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rating-system .

# Use a minimal alpine image for the final stage
FROM alpine:3.18

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/rating-system .

# Set the binary to be executable
RUN chmod +x /app/rating-system

# Use the non-root user
USER appuser

# Expose the application port
EXPOSE 8000

# Run the application
CMD ["/app/rating-system"]
