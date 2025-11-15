# Project variables
PROJECT_NAME := multiverse-core
DOCKER_IMAGE := $(PROJECT_NAME)

# Services
SERVICES := \
	entity-manager \
	narrative-orchestrator \
	world-generator \
	ban-of-world \
	city-governor \
	cultivation-module \
	reality-monitor \
	plan-manager \
	semantic-memory \
	ontological-archivist

# Default target
.PHONY: all
all: build

# Build all services
.PHONY: build
build:
	@echo "Building all services..."
	@docker build -t $(DOCKER_IMAGE) -f build/Dockerfile .

# Build specific service (make build-service SERVICE=entity-manager)
.PHONY: build-service
build-service:
	@echo "Building service: $(SERVICE)..."
	@CGO_ENABLED=0 GOOS=linux go build -o bin/$(SERVICE) ./cmd/$(SERVICE)

# Run services
.PHONY: up
up:
	@echo "Starting all services..."
	@docker-compose up -d

# Run specific service (make run SERVICE=entity-manager)
.PHONY: run
run:
	@echo "Starting service: $(SERVICE)..."
	@docker-compose up -d $(SERVICE)

# Stop services
.PHONY: down
down:
	@echo "Stopping all services..."
	@docker-compose down

# Clean
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@docker system prune -f

# Logs
.PHONY: logs
logs:
	@docker-compose logs -f

# Logs for specific service
.PHONY: logs-service
logs-service:
	@docker-compose logs -f $(SERVICE)

# Test (placeholder)
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Help
.PHONY: help
help:
	@echo "Makefile for $(PROJECT_NAME)"
	@echo ""
	@echo "Usage:"
	@echo "  make build              Build all services"
	@echo "  make build-service SERVICE=<name>  Build specific service"
	@echo "  make up                 Start all services"
	@echo "  make run SERVICE=<name> Start specific service"
	@echo "  make down               Stop all services"
	@echo "  make logs               Show logs for all services"
	@echo "  make logs-service SERVICE=<name>  Show logs for specific service"
	@echo "  make clean              Clean build artifacts"
	@echo "  make test               Run tests"
	@echo ""
	@echo "Available services:"
	@echo "  $(SERVICES)"