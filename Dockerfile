# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
# Clean install dependencies
RUN npm ci
COPY web/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
# Install build tools for CGO (sqlite)
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

# Stage 3: Final Runtime
FROM alpine:latest
WORKDIR /app
# Install runtime deps
RUN apk add --no-cache sqlite-libs ca-certificates tzdata

COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/web/dist ./dist
# Use example config as default
COPY config.example.yaml config.yaml

# Create a directory for persistent data (sqlite db)
RUN mkdir -p data

EXPOSE 8080
CMD ["./server"]
