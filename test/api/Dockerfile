# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY api/go.mod api/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code from api directory
COPY api/main.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gatling-api main.go

# Stage 2: Setup Maven environment for Gatling tests
FROM maven:3.9-eclipse-temurin-21-alpine

WORKDIR /app

# Install wget for healthcheck
RUN apk add --no-cache wget

# Install Go runtime to run the API
COPY --from=builder /app/gatling-api /app/gatling-api

# Copy the load-test project (adjust path based on build context)
COPY load-test /app/load-test

# Make sure mvnw is executable
RUN chmod +x /app/load-test/mvnw 2>/dev/null || true

# Create directory for reports
RUN mkdir -p /app/load-test/target/gatling

# Expose the API port
EXPOSE 8080

# Set working directory
WORKDIR /app

# Run the API
CMD ["/app/gatling-api"]
