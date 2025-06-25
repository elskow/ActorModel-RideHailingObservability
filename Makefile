# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Binary names
BINARY_NAME=actor-model-observability
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe

# Directories
CMD_DIR=./cmd
INTERNAL_DIR=./internal
SCRIPTS_DIR=./scripts
DOCS_DIR=./docs

# Docker
DOCKER_IMAGE=actor-model-observability
DOCKER_TAG=latest

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=actor_observability
DB_USER=postgres
DB_PASSWORD=postgres
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

.PHONY: all build clean test coverage deps fmt vet lint run dev docker help swagger docs

# Default target
all: clean deps fmt vet test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) -v $(CMD_DIR)

# Build for Linux
build-linux:
	@echo "Building $(BINARY_UNIX)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(CMD_DIR)

# Build for Windows
build-windows:
	@echo "Building $(BINARY_WINDOWS)..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) -v $(CMD_DIR)

# Build for all platforms
build-all: build build-linux build-windows

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_UNIX)
	@rm -f $(BINARY_WINDOWS)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Vet code
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install golangci-lint
install-lint:
	@echo "Installing golangci-lint..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run in development mode with live reload (requires air)
dev:
	@echo "Starting development server..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not installed. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run 'make install-dev-tools' to install development tools"; \
	fi

# Install development tools
install-dev-tools:
	@echo "Installing development tools..."
	$(GOCMD) install github.com/air-verse/air@latest
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs; \
	else \
		echo "Swag not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Alias for swagger target
docs: swagger

# Database operations
db-create:
	@echo "Creating database..."
	@createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || echo "Database might already exist"

db-drop:
	@echo "Dropping database..."
	@dropdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || echo "Database might not exist"

db-reset: db-drop db-create
	@echo "Database reset completed"

# Docker operations
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs:
	@echo "Showing docker-compose logs..."
	docker-compose logs -f

# Environment setup
setup-env:
	@echo "Setting up environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please update the .env file with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi

# Load testing framework
load-test:
	@echo "Running comprehensive load test..."
	@$(SCRIPTS_DIR)/run_load_test.sh

# Build load test tool
build-load-test:
	@echo "Building load test tool..."
	$(GOBUILD) -o load-test ./cmd/load-test

# Run quick load test (basic)
load-test-quick:
	@echo "Running quick load test..."
	@if command -v hey >/dev/null 2>&1; then \
		hey -n 1000 -c 10 http://localhost:8080/health; \
	elif command -v ab >/dev/null 2>&1; then \
		ab -n 1000 -c 10 http://localhost:8080/health; \
	else \
		echo "Load testing tool not found. Install hey: go install github.com/rakyll/hey@latest"; \
	fi

# Run Go benchmarks only
bench-comparison:
	@echo "Running comparison benchmarks..."
	$(GOTEST) -bench=BenchmarkActor -benchmem ./tests
	$(GOTEST) -bench=BenchmarkTraditional -benchmem ./tests
	$(GOTEST) -run=TestBenchmarkComparison ./tests

# Run specific benchmark tests
bench-actor:
	@echo "Running actor model benchmarks..."
	$(GOTEST) -bench=BenchmarkActor -benchmem ./tests

bench-traditional:
	@echo "Running traditional approach benchmarks..."
	$(GOTEST) -bench=BenchmarkTraditional -benchmem ./tests

bench-scalability:
	@echo "Running scalability benchmarks..."
	$(GOTEST) -bench=BenchmarkScalability -benchmem ./tests

bench-memory:
	@echo "Running memory usage benchmarks..."
	$(GOTEST) -bench=BenchmarkMemory -benchmem ./tests

bench-overhead:
	@echo "Running observability overhead benchmarks..."
	$(GOTEST) -bench=BenchmarkObservability -benchmem ./tests
	$(GOTEST) -bench=BenchmarkTraditionalMonitoring -benchmem ./tests

# Run benchmark comparison script
bench-script:
	@echo "Running benchmark comparison script..."
	$(GOBUILD) -o benchmark ./scripts
	./benchmark
	@rm -f benchmark

# Install load testing tools
install-load-tools:
	@echo "Installing load testing tools..."
	$(GOCMD) install github.com/rakyll/hey@latest
	@if command -v brew >/dev/null 2>&1; then \
		brew install apache-bench || echo "Apache Bench might already be installed"; \
	else \
		echo "Please install Apache Bench manually for your system"; \
	fi

# Security scan (requires gosec)
security-scan:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Install security tools
install-security-tools:
	@echo "Installing security tools..."
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Generate mocks (requires mockgen)
generate-mocks:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=$(INTERNAL_DIR)/repository/interfaces.go -destination=$(INTERNAL_DIR)/mocks/repository_mocks.go; \
	else \
		echo "mockgen not installed. Install it with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Install mock tools
install-mock-tools:
	@echo "Installing mock tools..."
	$(GOCMD) install github.com/golang/mock/mockgen@latest

# Check for updates
check-updates:
	@echo "Checking for dependency updates..."
	$(GOCMD) list -u -m all

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOCMD) get -u ./...
	$(GOMOD) tidy

# Verify dependencies
verify-deps:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Full CI pipeline
ci: deps fmt vet lint test coverage security-scan
	@echo "CI pipeline completed successfully"

# Development setup
setup: install-dev-tools install-security-tools install-mock-tools setup-env
	@echo "Development environment setup completed"
	@echo "Next steps:"
	@echo "1. Update .env file with your configuration"
	@echo "2. Start PostgreSQL and Redis services"
	@echo "3. Run 'make db-create' to create the database"
	@echo "4. Run 'make dev' to start the development server"

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  build-all          - Build for all platforms"
	@echo "  build-load-test    - Build load test tool"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  coverage           - Run tests with coverage"
	@echo "  test-race          - Run tests with race detection"
	@echo "  bench              - Run all benchmarks"
	@echo "  bench-comparison   - Run actor vs traditional comparison"
	@echo "  bench-actor        - Run actor model benchmarks"
	@echo "  bench-traditional  - Run traditional approach benchmarks"
	@echo "  bench-scalability  - Run scalability benchmarks"
	@echo "  bench-memory       - Run memory usage benchmarks"
	@echo "  bench-overhead     - Run observability overhead benchmarks"
	@echo "  bench-script       - Run benchmark comparison script"
	@echo "  load-test          - Run comprehensive load test framework"
	@echo "  load-test-quick    - Run quick load test"
	@echo "  deps               - Download dependencies"
	@echo "  fmt                - Format code"
	@echo "  vet                - Vet code"
	@echo "  lint               - Run linter"
	@echo "  run                - Run the application"
	@echo "  dev                - Run in development mode"
	@echo "  swagger            - Generate Swagger documentation"
	@echo "  db-create          - Create database"
	@echo "  db-drop            - Drop database"
	@echo "  db-reset           - Reset database"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  install-load-tools - Install load testing tools"
	@echo "  setup              - Setup development environment"
	@echo "  ci                 - Run CI pipeline"
	@echo "  help               - Show this help"