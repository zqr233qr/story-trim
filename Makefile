.PHONY: build run-server run-web clean wire generate test

# Build API server
build:
	go build -o bin/api-server ./cmd/api-server/

# Run API server
run-server:
	go run ./cmd/api-server/

# Generate Wire dependencies
wire:
	cd cmd/api-server && go generate

# Generate all code (includes wire)
generate:
	go generate ./...

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
