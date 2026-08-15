# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies first (better Docker caching)
COPY go.* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o app .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Run as a non-root user
RUN adduser -D appuser
USER appuser

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
