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
.PHONY: all build clean test lint docker-build docker-test docker-tidy fmt help mock-test

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
	@echo "  make fmt        - Format code using golangci-lint"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-test  - Run tests in Docker"
	@echo "  make docker-tidy - Run go mod tidy in Docker"
	@echo "  make shell      - Open a shell in the Docker container"
	@echo "  make mock-test  - Run commands with mock data"

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

mod:
	@echo "Running go mod..."
	docker run $(DOCKER_RUN_OPTS) \
		golang:$(GO_VERSION)-alpine \
		/bin/sh -c "go mod download && go mod verify && go mod tidy"

# Format/fix code using golangci-lint
lint:
	@echo "Formatting code..."
	docker run $(DOCKER_RUN_OPTS) \
		golangci/golangci-lint:latest \
		golangci-lint run --fix

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
		/bin/sh -c "go test -v ./..."

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

# Run commands with mock data
mock-test: build
	@echo "Running commands with mock data..."
	@echo "\nListing Resources:"
	./$(BINARY_NAME) list amis --mock
	./$(BINARY_NAME) list instances --mock
	./$(BINARY_NAME) list subnets --mock
	./$(BINARY_NAME) list keys --mock

	@echo "\nChecking Status:"
	./$(BINARY_NAME) check credentials --mock
	./$(BINARY_NAME) check migrate --check-instance-id i-1234567890abcdef0 --check-target-ami ami-0987654321fedcba0 --mock

	@echo "\nInstance Management:"
	./$(BINARY_NAME) create --mock --key test-key --subnet subnet-123 --ami ami-123 --type t2.micro --name test-instance
	./$(BINARY_NAME) backup --mock --instance-id i-1234567890abcdef0
	./$(BINARY_NAME) restore --mock --instance-id i-1234567890abcdef0 --snapshot snap-123
	./$(BINARY_NAME) restore --mock --instance-id i-1234567890abcdef0 --version v1.0.0

	@echo "\nInstance State Management:"
	./$(BINARY_NAME) start --mock --instance i-1234567890abcdef0
	./$(BINARY_NAME) stop --mock --instance i-1234567890abcdef0
	./$(BINARY_NAME) restart --mock --instance i-1234567890abcdef0

	@echo "\nMigration:"
	./$(BINARY_NAME) migrate --mock --instance-id i-1234567890abcdef0 --new-ami ami-0987654321fedcba0
	./$(BINARY_NAME) migrate --mock --enabled --new-ami ami-0987654321fedcba0
	./$(BINARY_NAME) migrate --mock --instance-id i-1234567890abcdef0 --version v1.0.0

	@echo "\nCleanup:"
	./$(BINARY_NAME) delete --mock --instance i-1234567890abcdef0

	@echo "\nAll mock tests completed successfully!"
