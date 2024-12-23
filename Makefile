# Variables
BINARY_NAME=ecman
DOCKER_IMAGE=ec-manager
GO_VERSION=1.21
GOCACHE=${HOME}/.cache/go-build
GOLANGCI=${HOME}/.cache/golangci-lint
GOMODCACHE=${HOME}/.cache/go-mod
GOSUMDB=${HOME}/.cache/go/sumdb
SHELL=/bin/bash
UID=$(shell id -u)
GID=$(shell id -g)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Use the latest git tag as the version, or default to dev
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")

# Add the date to the build information
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Docker run common options
DOCKER_RUN_OPTS=--rm \
	-v $(GOBASE):/app \
	-v $(GOLANGCI):/.cache/golangci-lint \
	-v $(GOCACHE):/.cache/go-build \
	-v $(GOMODCACHE):/go/pkg/mod \
	-v $(GOSUMDB):/go/pkg/sumdb \
	-w /app \
	--user $(UID):$(GID)

# PHONY targets
.PHONY: all build clean test lint docker-build docker-test docker-tidy fmt help

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
	@echo "  make fmt        - Format code using goimports"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-test  - Run tests in Docker"
	@echo "  make docker-tidy - Run go mod tidy in Docker"
	@echo "  make shell      - Open a shell in the Docker container"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		go build -o $(BINARY_NAME) \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Interactive shell
shell:
	docker run -it $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		/bin/sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf bin/

# Run tests
test:
	@echo "Running tests..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		/bin/sh -c "go test -v ./... -count=1"

# Run linter
lint:
	@echo "Running linter..."
	docker run $(DOCKER_RUN_OPTS) \
		golangci/golangci-lint:latest \
		golangci-lint run ./...

# Format code using goimports
fmt:
	@echo "Formatting code..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		/bin/sh -c "go install golang.org/x/tools/cmd/goimports@v0.16.1 && \
		find . -type f -name '*.go' ! -path './vendor/*' -exec goimports -w {} \;"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Run tests in Docker
docker-test:
	@echo "Running tests in Docker..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		/bin/sh -c "go mod download && go test -v ./..."

# Run go mod tidy in Docker
docker-tidy:
	@echo "Running go mod tidy..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		go mod tidy

# Create go.mod if it doesn't exist
init:
	@if [ ! -f go.mod ]; then \
		echo "Initializing go.mod..."; \
		docker run $(DOCKER_RUN_OPTS) \
			golang:$(GO_VERSION)-alpine \
			go mod init github.com/taemon1337/ami-migrate; \
	fi
