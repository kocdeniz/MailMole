# MailMole Dockerfile
# Multi-stage build for minimal image size

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mailmole .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -u 1000 mailmole

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mailmole .

# Create directories
RUN mkdir -p /data/logs /data/exports /data/backups && \
    chown -R mailmole:mailmole /data

# Switch to non-root user
USER mailmole

# Expose port for web dashboard
EXPOSE 8080

# Volume for persistent data
VOLUME ["/data"]

# Default command
CMD ["./mailmole", "-web", ":8080", "-web-only"]
