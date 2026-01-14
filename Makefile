.PHONY: help build clean test lint docker-build docker-push helm-lint helm-package run

# Variables
BINARY_NAME=traefik-updater
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
DOCKER_REPO?=tazhate/k8s-internal-loadbalancer
HELM_CHART=./chart

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

build-local: ## Build for local OS
	@echo "Building $(BINARY_NAME) for local OS..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin" && exit 1)
	golangci-lint run --timeout=5m
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Formatting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

mod-download: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

mod-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied"

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -t $(DOCKER_REPO):$(VERSION) -f Dockerfile .
	docker build -t $(DOCKER_REPO)-traefik:$(VERSION) -f Dockerfile-traefik .
	@echo "Docker images built"

docker-build-latest: docker-build ## Build and tag as latest
	@echo "Tagging as latest..."
	docker tag $(DOCKER_REPO):$(VERSION) $(DOCKER_REPO):latest
	docker tag $(DOCKER_REPO)-traefik:$(VERSION) $(DOCKER_REPO)-traefik:latest
	@echo "Tagged as latest"

docker-push: ## Push Docker images
	@echo "Pushing Docker images..."
	docker push $(DOCKER_REPO):$(VERSION)
	docker push $(DOCKER_REPO)-traefik:$(VERSION)
	@echo "Docker images pushed"

docker-push-latest: ## Push latest tags
	@echo "Pushing latest tags..."
	docker push $(DOCKER_REPO):latest
	docker push $(DOCKER_REPO)-traefik:latest
	@echo "Latest tags pushed"

helm-lint: ## Lint Helm chart
	@echo "Linting Helm chart..."
	helm lint $(HELM_CHART)
	@echo "Helm chart lint complete"

helm-template: ## Template Helm chart
	@echo "Templating Helm chart..."
	helm template test $(HELM_CHART) --debug
	@echo "Helm template complete"

helm-package: ## Package Helm chart
	@echo "Packaging Helm chart..."
	helm package $(HELM_CHART)
	@echo "Helm chart packaged"

run: ## Run locally (requires kubeconfig)
	@echo "Running $(BINARY_NAME) locally..."
	@if [ -z "$$POD_LABELS" ]; then echo "Error: POD_LABELS not set"; exit 1; fi
	@if [ -z "$$TRAEFIK_API_URL" ]; then echo "Error: TRAEFIK_API_URL not set"; exit 1; fi
	@if [ -z "$$POD_NAMESPACE" ]; then echo "Error: POD_NAMESPACE not set"; exit 1; fi
	./$(BINARY_NAME)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Tools installed"

all: clean lint test build ## Clean, lint, test, and build

release: clean lint test build docker-build docker-push ## Full release pipeline

.DEFAULT_GOAL := help
