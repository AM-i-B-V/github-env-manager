# GitHub Environment Manager Makefile

.PHONY: help build run test clean deps lint format docker-build docker-run build-all release version release-patch release-minor release-major

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application for current platform"
	@echo "  build-all     - Build for all platforms (Linux, macOS, Windows)"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  release       - Create release binaries"
	@echo "  version       - Show current version"
	@echo "  release-patch - Create patch release (0.0.x)"
	@echo "  release-minor - Create minor release (0.x.0)"
	@echo "  release-major - Create major release (x.0.0)"

# Build the application for current platform
build:
	@echo "Building GitHub Environment Manager..."
	go build -ldflags="-s -w" -o bin/github-env-manager cmd/server/main.go

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/github-env-manager-linux-amd64 cmd/server/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/github-env-manager-linux-arm64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/github-env-manager-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/github-env-manager-darwin-arm64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/github-env-manager-windows-amd64.exe cmd/server/main.go
	@echo "Build complete! Binaries are in the bin/ directory"

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@mkdir -p releases
	cd bin && tar -czf ../releases/github-env-manager_linux_amd64.tar.gz github-env-manager-linux-amd64
	cd bin && tar -czf ../releases/github-env-manager_linux_arm64.tar.gz github-env-manager-linux-arm64
	cd bin && tar -czf ../releases/github-env-manager_darwin_amd64.tar.gz github-env-manager-darwin-amd64
	cd bin && tar -czf ../releases/github-env-manager_darwin_arm64.tar.gz github-env-manager-darwin-arm64
	cd bin && zip ../releases/github-env-manager_windows_amd64.zip github-env-manager-windows-amd64.exe
	@echo "Release archives created in releases/ directory"

# Version management
version:
	@chmod +x scripts/version.sh
	@./scripts/version.sh version

release-patch:
	@chmod +x scripts/version.sh
	@./scripts/version.sh release auto patch

release-minor:
	@chmod +x scripts/version.sh
	@./scripts/version.sh release auto minor

release-major:
	@chmod +x scripts/version.sh
	@./scripts/version.sh release auto major

# Run the application
run:
	@echo "Running GitHub Environment Manager..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf releases/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t github-env-manager .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 github-env-manager

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	# TODO: Add swagger generation

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Performance benchmark
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Install script permissions
install-scripts:
	@echo "Setting up install scripts..."
	chmod +x scripts/install.sh
	chmod +x scripts/version.sh
	@echo "Install scripts are ready!"

# Quick install (for development)
quick-install: build
	@echo "Installing locally..."
	@mkdir -p ~/.local/bin
	cp bin/github-env-manager ~/.local/bin/
	@echo "Installed to ~/.local/bin/github-env-manager"
	@echo "Add ~/.local/bin to your PATH if not already there"
