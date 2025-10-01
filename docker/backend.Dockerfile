# Backend Dockerfile for VolunteerSync Go API
# Multi-stage build for optimal image size and security

# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the entire backend source code
COPY backend/ ./

# Build the application
# CGO_ENABLED=0 for fully static binary
# -ldflags for smaller binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /build/api \
    ./cmd/api

# Stage 2: Production stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/api /app/api

# Copy migrations directory (needed for database migrations)
COPY --from=builder /build/migrations /app/migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/api"]
