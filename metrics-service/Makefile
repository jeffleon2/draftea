.PHONY: test test-coverage test-verbose clean docker-up docker-down docker-logs docker-restart

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose:
	go test ./... -v -count=1

# Clean generated files
clean:
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download

# Run the service
run:
	go run cmd/main.go

# Build the service
build:
	go build -o bin/metrics-service cmd/main.go

# Docker commands
docker-up:
	docker-compose up -d

# Stop docker containers
docker-down:
	docker-compose down

# View docker logs
docker-logs:
	docker-compose logs -f

# Restart docker containers
docker-restart:
	docker-compose restart

# Rebuild and start containers
docker-rebuild:
	docker-compose up -d --build
