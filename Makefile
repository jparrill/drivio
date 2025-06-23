# Variables
BINARY_NAME=drivio
BUILD_DIR=build
VERSION?=0.1.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.CommitHash=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

.PHONY: all build clean test deps lint help install uninstall release release-snapshot validate-release

# Default target
all: clean build

# Build the application
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	$(GOBUILD) ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf ${BUILD_DIR}
	@rm -rf dist

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Lint the code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install the binary to GOPATH
install: build
	@echo "Installing ${BINARY_NAME}..."
	@cp ${BUILD_DIR}/${BINARY_NAME} $(shell go env GOPATH)/bin/

# Uninstall the binary
uninstall:
	@echo "Uninstalling ${BINARY_NAME}..."
	@rm -f $(shell go env GOPATH)/bin/${BINARY_NAME}

# Run the application
run:
	@echo "Running ${BINARY_NAME}..."
	$(GOCMD) run .

# Development mode with hot reload (requires air)
dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run with: go run ."; \
		$(GOCMD) run .; \
	fi

# Release management with goreleaser
release:
	@echo "Creating release with goreleaser..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not found. Install it with: go install github.com/goreleaser/goreleaser/v2/cmd/goreleaser@latest"; \
	fi

release-snapshot:
	@echo "Creating snapshot release with goreleaser..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean --skip-publish; \
	else \
		echo "goreleaser not found. Install it with: go install github.com/goreleaser/goreleaser/v2/cmd/goreleaser@latest"; \
	fi

# Validate goreleaser configuration
validate-release:
	@echo "Validating goreleaser configuration..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "goreleaser not found. Install it with: go install github.com/goreleaser/goreleaser/v2/cmd/goreleaser@latest"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Lint the code"
	@echo "  install      - Install the binary"
	@echo "  uninstall    - Uninstall the binary"
	@echo "  run          - Run the application"
	@echo "  dev          - Run in development mode"
	@echo "  release      - Create a release with goreleaser"
	@echo "  release-snapshot - Create a snapshot release"
	@echo "  validate-release - Validate goreleaser config"
	@echo "  help         - Show this help"