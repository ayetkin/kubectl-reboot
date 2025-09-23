# Variables
PLUGIN_NAME := kubectl-reboot
VERSION := v1.0.0
PACKAGE := github.com/ayetkin/kubectl-reboot
MAIN_PATH := cmd/k8s-restart

# Build information
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)

# Platforms for release
PLATFORMS := \
    linux/amd64 \
    linux/arm64 \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64

.PHONY: help build clean test release install-local krew-manifest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the plugin binary
	@echo "Building $(PLUGIN_NAME)..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(PLUGIN_NAME) ./$(MAIN_PATH)

clean: ## Remove build artifacts
	@echo "Cleaning up..."
	@rm -rf bin/ dist/

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

tidy: ## Tidy up go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

release: clean ## Build release binaries for all platforms
	@echo "Building release binaries..."
	@mkdir -p dist
	@$(foreach platform,$(PLATFORMS), \
		echo "Building for $(platform)..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 \
		go build -ldflags "$(LDFLAGS)" \
		-o dist/$(PLUGIN_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(findstring windows,$(platform)),.exe,) \
		./$(MAIN_PATH); \
	)

package: release ## Package release binaries
	@echo "Packaging release binaries..."
	@for file in dist/$(PLUGIN_NAME)-*; do \
		if [[ "$$file" == *".exe" ]]; then \
			platform=$$(basename $$file .exe | sed 's/$(PLUGIN_NAME)-//'); \
			zip -j dist/$(PLUGIN_NAME)-$$platform.zip $$file; \
		else \
			platform=$$(basename $$file | sed 's/$(PLUGIN_NAME)-//'); \
			tar -czf dist/$(PLUGIN_NAME)-$$platform.tar.gz -C dist $$(basename $$file); \
		fi; \
	done

install-local: build ## Install the plugin locally
	@echo "Installing $(PLUGIN_NAME) locally..."
	@mkdir -p ~/.krew/bin
	@cp bin/$(PLUGIN_NAME) ~/.krew/bin/
	@echo "Plugin installed to ~/.krew/bin/$(PLUGIN_NAME)"
	@echo "Make sure ~/.krew/bin is in your PATH"

krew-manifest: ## Generate checksums for Krew manifest
	@echo "Generating checksums for Krew manifest..."
	@if [ ! -d "dist" ]; then echo "Run 'make package' first"; exit 1; fi
	@for file in dist/*.tar.gz dist/*.zip; do \
		if [ -f "$$file" ]; then \
			echo "$$file: $$(sha256sum $$file | cut -d' ' -f1)"; \
		fi; \
	done

check: vet fmt test ## Run all checks

all: check build ## Run checks and build
