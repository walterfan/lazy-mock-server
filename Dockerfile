# Multi-stage Dockerfile for Lazy Mock Server

# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY main.go ./

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o mock-server main.go

# Final stage - minimal image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /build/mock-server /app/mock-server

# Copy configuration files
COPY app/ /app/config/app/
COPY config/ /app/config/config/

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/mock-server", "-config", "/app/config/app/mock_response.yaml", "-port", "8080"]

# Set working directory
WORKDIR /app

# Default command
ENTRYPOINT ["/app/mock-server"]
CMD ["-config", "/app/config/app/mock_response.yaml", "-port", "8080"]
