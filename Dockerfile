# Multi-stage build for Malice - Ubuntu 22.04 compatible
# Stage 1: Builder
FROM golang:1.21-alpine AS builder

ARG VERSION=0.3.28
ARG COMMIT=unknown
ARG BUILD_DATE

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.date=${BUILD_DATE}" \
    -o malice .

# Stage 2: Runtime
FROM ubuntu:22.04

LABEL maintainer="blacktop <https://github.com/blacktop>"
LABEL description="Open Source Malware Analysis Framework"
LABEL version="0.3.28"

# Install runtime dependencies for Ubuntu 22.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    docker.io \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /malice

# Copy binary from builder
COPY --from=builder /app/malice /usr/local/bin/malice

# Create directories
RUN mkdir -p /malice/samples /malice/config /malice/logs

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD malice --help > /dev/null || exit 1

ENTRYPOINT ["malice"]
CMD ["--help"]
