.PHONY: build test integration-test docker-up docker-down clean dev

# Build the Go binary
build:
	go build -o bin/continuum ./cmd/continuum

# Run unit tests
test:
	go test -v ./...

# Run integration tests (requires docker-compose up)
integration-test:
	go test -v ./tests/integration/... -count=1

# Start all services with docker-compose
docker-up:
	docker-compose up --build -d

# Stop all services
docker-down:
	docker-compose down -v

# Clean build artifacts
clean:
	rm -rf bin/ dist/
	go clean

# Run in development mode
dev:
	go run ./cmd/continuum

# Install frontend dependencies
web-install:
	cd web && npm install

# Build frontend
web-build:
	cd web && npm run build

# Run frontend dev server
web-dev:
	cd web && npm run dev
