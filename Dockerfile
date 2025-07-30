# Multi-stage build for AWS Transformer CLI Tool
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Create final lightweight image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 transformer && \
    adduser -D -s /bin/bash -u 1000 -G transformer transformer

# Set working directory
WORKDIR /workspace

# Copy binary from builder stage
COPY --from=builder /app/build/transformer /usr/local/bin/transformer

# Make binary executable
RUN chmod +x /usr/local/bin/transformer

# Switch to non-root user
USER transformer

# Set default command
ENTRYPOINT ["transformer"]

# Default command (can be overridden)
CMD ["--help"] 