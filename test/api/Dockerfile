# Stage 1: Build Go API
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o gatling-api main.go

# Stage 2: Runtime with Maven (non-alpine for native libs)
FROM maven:3.9-eclipse-temurin-21
WORKDIR /app

# Install wget for healthcheck
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

# Copy Go API
COPY --from=builder /app/gatling-api /app/gatling-api

# Copy load-test project
COPY load-test /app/load-test

# Make Maven wrapper executable
RUN chmod +x /app/load-test/mvnw

# Pre-download dependencies, plugins and compile tests
WORKDIR /app/load-test
RUN mvn dependency:resolve dependency:resolve-plugins -B
RUN mvn clean test-compile -DskipTests -B
# Dry-run to force plugin dependencies download
RUN mvn gatling:help -B

# Create reports directory
RUN mkdir -p /app/load-test/target/gatling

EXPOSE 8080
WORKDIR /app
CMD ["/app/gatling-api"]
