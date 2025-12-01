FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o gatling-api main.go

FROM maven:3.9-eclipse-temurin-21-alpine
WORKDIR /app
RUN apk add --no-cache wget
COPY --from=builder /app/gatling-api /app/gatling-api
COPY load-test /app/load-test
RUN chmod +x /app/load-test/mvnw 2>/dev/null || true
WORKDIR /app/load-test
RUN ./mvnw dependency:go-offline -B || true
RUN mkdir -p /app/load-test/target/gatling
EXPOSE 8080
WORKDIR /app
CMD ["/app/gatling-api"]
