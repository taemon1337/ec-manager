# Variables
BINARY_NAME=ami-migrate
DOCKER_IMAGE=ami-migrate
GO_VERSION=1.21
SHELL=/bin/bash

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Use the latest git tag as the version, or default to dev
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")

# Add the date to the build information
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# PHONY targets
.PHONY: all build clean test lint docker-build docker-test help

# Default target
all: clean build test

# Help target
help:
	@echo "Available targets:"
	@echo "  make all         - Clean, build, and test"
	@echo "  make build      - Build the binary"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests in Docker"
	@echo "  make lint       - Run linter in Docker"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-test  - Run tests in Docker"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	docker run --rm \
		-v $(GOBASE):/app \
		-w /app \
		golang:$(GO_VERSION)-alpine \
		go build -o $(BINARY_NAME) \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf bin/

# Run tests
test:
	@echo "Running tests..."
	docker run --rm \
		-v $(GOBASE):/app \
		-w /app \
		golang:$(GO_VERSION)-alpine \
		go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	docker run --rm \
		-v $(GOBASE):/app \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Run tests in Docker
docker-test:
	@echo "Running tests in Docker..."
	docker run --rm \
		-v $(GOBASE):/app \
		-w /app \
		golang:$(GO_VERSION)-alpine \
		/bin/sh -c "go test -v ./..."

# Create go.mod if it doesn't exist
init:
	@if [ ! -f go.mod ]; then \
		echo "Initializing go.mod..."; \
		docker run --rm \
			-v $(GOBASE):/app \
			-w /app \
			golang:$(GO_VERSION)-alpine \
			go mod init github.com/taemon1337/ami-migrate; \
	fi
