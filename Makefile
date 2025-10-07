# Screenshot MCP Server Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build information
BINARY_NAME_SERVER=screenshot-server
BINARY_NAME_CLI=mcpctl
BINARY_DIR=bin
PKG_DIR=cmd

# Version information
VERSION?=1.0.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE) -s -w"

# Default target
.PHONY: all
all: clean deps build

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build for current OS
.PHONY: build
build: build-server build-cli

.PHONY: build-server
build-server:
	@echo "Building server for current OS..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME_SERVER).exe ./$(PKG_DIR)/server

.PHONY: build-cli
build-cli:
	@echo "Building CLI for current OS..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME_CLI).exe ./$(PKG_DIR)/mcpctl

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)/windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/windows/$(BINARY_NAME_SERVER).exe ./$(PKG_DIR)/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/windows/$(BINARY_NAME_CLI).exe ./$(PKG_DIR)/mcpctl

# Build for all platforms
.PHONY: build-all
build-all: build-windows build-linux build-darwin

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/linux/$(BINARY_NAME_SERVER) ./$(PKG_DIR)/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/linux/$(BINARY_NAME_CLI) ./$(PKG_DIR)/mcpctl

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)/darwin
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/darwin/$(BINARY_NAME_SERVER) ./$(PKG_DIR)/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/darwin/$(BINARY_NAME_CLI) ./$(PKG_DIR)/mcpctl

# Run the server
.PHONY: run-server
run-server: build-server
	@echo "Starting screenshot server..."
	./$(BINARY_DIR)/$(BINARY_NAME_SERVER).exe --port 8080

# Run the CLI
.PHONY: run-cli
run-cli: build-cli
	@echo "Running CLI..."
	./$(BINARY_DIR)/$(BINARY_NAME_CLI).exe --help

# Test the application
.PHONY: test test-unit test-integration test-all-tests
test: test-unit
	@echo "Unit tests completed"

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./pkg/... ./internal/...

test-integration:
	@echo "Running integration tests..."
	cd test && $(GOMOD) tidy && $(GOTEST) -v -timeout 60s ./...

test-all-tests: test-unit test-integration
	@echo "All tests completed"

# Test with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmark tests
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linters
.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "# Screenshot MCP Server Documentation" > docs/index.md
	@echo "" >> docs/index.md
	@echo "Generated on: $(BUILD_DATE)" >> docs/index.md
	@echo "Version: $(VERSION)" >> docs/index.md
	@echo "Commit: $(GIT_COMMIT)" >> docs/index.md

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t screenshot-mcp-server:$(VERSION) .

# Docker run
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 screenshot-mcp-server:$(VERSION)

# Install binaries to system
.PHONY: install
install: build
	@echo "Installing binaries..."
	@mkdir -p $(HOME)/bin
	cp $(BINARY_DIR)/$(BINARY_NAME_SERVER).exe $(HOME)/bin/
	cp $(BINARY_DIR)/$(BINARY_NAME_CLI).exe $(HOME)/bin/
	@echo "Binaries installed to $(HOME)/bin/"

# Uninstall binaries
.PHONY: uninstall
uninstall:
	@echo "Uninstalling binaries..."
	rm -f $(HOME)/bin/$(BINARY_NAME_SERVER).exe
	rm -f $(HOME)/bin/$(BINARY_NAME_CLI).exe

# Development setup
.PHONY: dev-setup
dev-setup: install-tools deps
	@echo "Development environment setup complete"

# Quick test - build and run basic functionality
.PHONY: quick-test
quick-test: build-cli
	@echo "Running quick functionality test..."
	@echo "Testing Chrome instance discovery..."
	./$(BINARY_DIR)/$(BINARY_NAME_CLI).exe chrome instances || echo "Chrome discovery failed (may be expected if no Chrome running)"
	@echo "Quick test completed"

# Release build with optimizations
.PHONY: release
release: clean deps test
	@echo "Building release binaries..."
	$(MAKE) build-all
	@echo "Release build completed"

# Help
.PHONY: help
help:
	@echo "Screenshot MCP Server - Makefile Help"
	@echo ""
	@echo "Available targets:"
	@echo "  all          - Clean, download deps, and build"
	@echo "  deps         - Download Go dependencies"
	@echo "  build        - Build server and CLI for current OS"
	@echo "  build-server - Build only the server"
	@echo "  build-cli    - Build only the CLI"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  run-server   - Build and run the server"
	@echo "  run-cli      - Build and run the CLI"
	@echo "  test         - Run unit tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all-tests - Run all tests (unit + integration)"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  bench        - Run benchmark tests"
	@echo "  clean        - Remove build artifacts"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Run code linters"
	@echo "  docs         - Generate documentation"
	@echo "  install      - Install binaries to ~/bin"
	@echo "  uninstall    - Remove installed binaries"
	@echo "  dev-setup    - Set up development environment"
	@echo "  quick-test   - Quick functionality test"
	@echo "  release      - Build optimized release binaries"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build for current OS"
	@echo "  make run-server              # Start the server"
	@echo "  make quick-test              # Test basic functionality"
	@echo "  make VERSION=1.1.0 release  # Build release with custom version"