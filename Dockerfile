# Multi-stage build to reduce image size
FROM golang:1.25-alpine AS builder

# Install git for go modules
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
RUN CGO_ENABLED=0 GOOS=linux go build -o tronevent-daemon ./cmd/daemon/main.go

# Final stage
FROM alpine:3.22.1

# Install ca-certificates for HTTPS requests and create non-root user
RUN apk --no-cache add ca-certificates \
    && update-ca-certificates \
    && adduser -D -s /bin/sh tronevent

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/tronevent-daemon /usr/local/bin/tronevent-daemon

# Create directory for config and state files
RUN mkdir -p /app/config && chown -R tronevent:tronevent /app

# Switch to non-root user
USER tronevent

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD pgrep tronevent-daemon || exit 1

# Command to run the binary
CMD ["tronevent-daemon"]