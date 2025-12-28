##############################
# VARIABLES
##############################
# Load .env if exists (loaded first so defaults can override if not set)
-include .env

# Default values (can be overridden by .env or command line)
APP_NAME ?= glcron
APP_VERSION ?= 0.0.1-local
CONTAINER_IMAGE_NAME ?= glcron
BUILD_GOVERSION ?= 1.23
BUILD_GOOS ?= $(shell go env GOOS)
BUILD_GOARCH ?= $(shell go env GOARCH)

# Build flags
LDFLAGS := -X glcron/internal/tui.AppVersion=$(APP_VERSION)

##############################
# HELP
##############################
.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Variables (from .env or defaults):"
	@echo "  APP_NAME=$(APP_NAME)"
	@echo "  APP_VERSION=$(APP_VERSION)"
	@echo "  BUILD_GOOS=$(BUILD_GOOS)"
	@echo "  BUILD_GOARCH=$(BUILD_GOARCH)"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help

##############################
# BUILD
##############################
.PHONY: build
build: ## Build the application binary
	@echo "Building $(APP_NAME) v$(APP_VERSION) for $(BUILD_GOOS)/$(BUILD_GOARCH)..."
	@go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) ./cmd/$(APP_NAME)
	@echo "✓ Built ./$(APP_NAME)"

.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building $(APP_NAME) v$(APP_VERSION) for multiple platforms..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-darwin-amd64 ./cmd/$(APP_NAME)
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-darwin-arm64 ./cmd/$(APP_NAME)
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux-arm64 ./cmd/$(APP_NAME)
	@echo "✓ Built binaries in ./dist/"

.PHONY: run
run: build ## Build and run the application
	@./$(APP_NAME)

.PHONY: dev
dev: ## Run the application in development mode
	@go run -ldflags "$(LDFLAGS)" ./cmd/$(APP_NAME)

.PHONY: clean
clean: ## Clean build artifacts
	@rm -f $(APP_NAME)
	@rm -rf dist/
	@echo "✓ Cleaned build artifacts"

##############################
# QUALITY
##############################
.PHONY: fmt
fmt: ## Format code
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

.PHONY: test
test: ## Run tests
	@go test -v ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@golangci-lint run

##############################
# UTILITY
##############################
.PHONY: install
install: build ## Install binary to $GOPATH/bin
	@go install -ldflags "$(LDFLAGS)" ./cmd/$(APP_NAME)
	@echo "✓ Installed $(APP_NAME) to $(shell go env GOPATH)/bin/"

.PHONY: deps
deps: ## Download and tidy dependencies
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

.PHONY: version
version: ## Show version info
	@echo "App: $(APP_NAME)"
	@echo "Version: $(APP_VERSION)"
	@echo "Go: $(shell go version)"
	@echo "OS/Arch: $(BUILD_GOOS)/$(BUILD_GOARCH)"
