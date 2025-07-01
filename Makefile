# Makefile for Terraform Pocket-ID Provider

# Variables
PROVIDER_NAME = pocketid
NAMESPACE = trozz
BINARY_NAME = terraform-provider-$(PROVIDER_NAME)
VERSION ?= 0.1.0
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_PATH = ~/.terraform.d/plugins/registry.terraform.io/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(OS_ARCH)

# Go variables
GOTEST = go test
GOBUILD = go build
GOFMT = gofmt
GOVET = go vet
GOMOD = go mod
GOLINT = golangci-lint

# Terraform variables
TF_LOG ?=
TF_ACC ?=

# Colors for output
CYAN = \033[0;36m
GREEN = \033[0;32m
RED = \033[0;31m
YELLOW = \033[0;33m
NC = \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Terraform Pocket-ID Provider Makefile$(NC)"
	@echo "$(GREEN)Usage:$(NC) make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the provider binary
	@echo "$(GREEN)Building provider...$(NC)"
	$(GOBUILD) -o $(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

.PHONY: install
install: build ## Build and install the provider locally
	@echo "$(GREEN)Installing provider to $(INSTALL_PATH)...$(NC)"
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "$(GREEN)Provider installed successfully!$(NC)"

.PHONY: test
test: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) -v -cover -coverprofile=coverage.out ./internal/...
	@echo "$(GREEN)Unit tests complete!$(NC)"

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage report
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-acc
test-acc: ## Run acceptance tests (requires POCKETID_BASE_URL and POCKETID_API_TOKEN)
	@if [ -z "$(POCKETID_BASE_URL)" ]; then \
		echo "$(RED)Error: POCKETID_BASE_URL environment variable is not set$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$(POCKETID_API_TOKEN)" ]; then \
		echo "$(RED)Error: POCKETID_API_TOKEN environment variable is not set$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running acceptance tests...$(NC)"
	TF_ACC=1 $(GOTEST) -v -timeout 30m ./internal/... -tags=acc
	@echo "$(GREEN)Acceptance tests complete!$(NC)"

.PHONY: test-all
test-all: test test-acc ## Run all tests

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOFMT) -w -s .
	@echo "$(GREEN)Code formatted!$(NC)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)The following files need formatting:$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	else \
		echo "$(GREEN)All files are properly formatted!$(NC)"; \
	fi

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install it with:$(NC)"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOVET) ./...
	@echo "$(GREEN)go vet complete!$(NC)"

.PHONY: check
check: fmt-check vet lint test ## Run all checks (format, vet, lint, test)
	@echo "$(GREEN)All checks passed!$(NC)"

.PHONY: docs
docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	@if command -v tfplugindocs >/dev/null 2>&1; then \
		tfplugindocs generate; \
		echo "$(GREEN)Documentation generated!$(NC)"; \
	else \
		echo "$(YELLOW)tfplugindocs not installed. Install it with:$(NC)"; \
		echo "  go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
	fi

.PHONY: docs-preview
docs-preview: ## Preview documentation in browser
	@echo "$(GREEN)Starting documentation preview server...$(NC)"
	@cd docs && python3 -m http.server 8000

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "$(GREEN)Clean complete!$(NC)"

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

.PHONY: update-deps
update-deps: ## Update all dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@$(GOMOD) get -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

.PHONY: dev
dev: install ## Build and install for development
	@echo "$(GREEN)Development build installed!$(NC)"
	@echo "$(CYAN)You can now use the provider in your Terraform configurations$(NC)"

.PHONY: release-dry-run
release-dry-run: ## Run goreleaser in dry-run mode
	@echo "$(GREEN)Running release dry-run...$(NC)"
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
		echo "$(GREEN)Release dry-run complete! Check dist/ directory$(NC)"; \
	else \
		echo "$(YELLOW)goreleaser not installed. Install it with:$(NC)"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
	fi

.PHONY: release
release: ## Create a new release (requires GITHUB_TOKEN)
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "$(RED)Error: GITHUB_TOKEN environment variable is not set$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating release...$(NC)"
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "$(YELLOW)goreleaser not installed. Install it with:$(NC)"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
	fi

.PHONY: example-init
example-init: ## Initialize example configurations
	@echo "$(GREEN)Initializing examples...$(NC)"
	@cd examples/complete && terraform init
	@cd examples/provider && terraform init
	@cd examples/resources && terraform init
	@echo "$(GREEN)Examples initialized!$(NC)"

.PHONY: example-plan
example-plan: ## Run terraform plan on complete example
	@echo "$(GREEN)Running terraform plan on complete example...$(NC)"
	@cd examples/complete && terraform plan

.PHONY: test-integration
test-integration: ## Run integration tests against live Pocket-ID instance
	@echo "$(GREEN)Running integration tests...$(NC)"
	@cd test && terraform init && terraform apply -auto-approve
	@echo "$(GREEN)Integration tests complete!$(NC)"

.PHONY: test-cleanup
test-cleanup: ## Clean up integration test resources
	@echo "$(GREEN)Cleaning up test resources...$(NC)"
	@cd test && terraform destroy -auto-approve
	@echo "$(GREEN)Test cleanup complete!$(NC)"

.PHONY: setup-hooks
setup-hooks: ## Set up git hooks
	@echo "$(GREEN)Setting up git hooks...$(NC)"
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make fmt-check" >> .git/hooks/pre-commit
	@echo "make vet" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)Git hooks installed!$(NC)"

.PHONY: docker-test
docker-test: ## Run tests in Docker container
	@echo "$(GREEN)Running tests in Docker...$(NC)"
	@docker run --rm -v $(PWD):/workspace -w /workspace golang:1.21 make test

.PHONY: pocket-id-start
pocket-id-start: ## Start local Pocket-ID instance using Docker
	@echo "$(GREEN)Starting Pocket-ID instance...$(NC)"
	@cd pocket-id-source && docker-compose up -d
	@echo "$(GREEN)Pocket-ID is running at http://localhost:8080$(NC)"

.PHONY: pocket-id-stop
pocket-id-stop: ## Stop local Pocket-ID instance
	@echo "$(GREEN)Stopping Pocket-ID instance...$(NC)"
	@cd pocket-id-source && docker-compose down
	@echo "$(GREEN)Pocket-ID stopped$(NC)"

.PHONY: pocket-id-logs
pocket-id-logs: ## Show Pocket-ID logs
	@cd pocket-id-source && docker-compose logs -f

# Debug helpers
.PHONY: debug-env
debug-env: ## Show environment variables for debugging
	@echo "$(CYAN)Environment Variables:$(NC)"
	@echo "  POCKETID_BASE_URL  = $(POCKETID_BASE_URL)"
	@echo "  POCKETID_API_TOKEN = $(if $(POCKETID_API_TOKEN),[SET],[NOT SET])"
	@echo "  TF_LOG             = $(TF_LOG)"
	@echo "  TF_ACC             = $(TF_ACC)"
	@echo "  OS_ARCH            = $(OS_ARCH)"
	@echo "  VERSION            = $(VERSION)"

.PHONY: version
version: ## Show version information
	@echo "$(CYAN)Terraform Pocket-ID Provider$(NC)"
	@echo "  Version: $(VERSION)"
	@echo "  Go Version: $(shell go version)"
	@echo "  Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Built: $(shell date)"

# Default target
.DEFAULT_GOAL := help
