# Build Stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies (CGO requires gcc)
RUN apk add --no-cache build-base

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (strip debug info)
# Ensure the path matches your project structure: cmd/api-server/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o api-server ./cmd/api-server

# Runtime Stage
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs ca-certificates tzdata

# Copy binary
COPY --from=builder /app/api-server .

# Copy example config as default
COPY config.example.yaml config.yaml

# Create directories for persistence
RUN mkdir -p uploads data

# Expose API port
EXPOSE 8080

# Start server
CMD ["./api-server"]