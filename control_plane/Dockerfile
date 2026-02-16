# Multi-stage build for control plane
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod file from root
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy control plane source code
COPY control_plane/ ./control_plane/

# Build the application
WORKDIR /app/control_plane
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fluxforge-control-plane .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Copy binary from builder
COPY --from=builder /app/control_plane/fluxforge-control-plane .

# Create non-root user
RUN addgroup -g 1000 fluxforge && \
    adduser -D -u 1000 -G fluxforge fluxforge && \
    chown -R fluxforge:fluxforge /app

USER fluxforge

# Expose ports
EXPOSE 8080 9000

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["./fluxforge-control-plane"]
