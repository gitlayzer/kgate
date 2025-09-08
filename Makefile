# Makefile for the kgate project

# --- Variables ---
# The name of your application's binary
BINARY_NAME=kgate

# The directory to place build artifacts
BUILD_DIR=./bin

# --- Dynamic Variables ---
# Use git describe to get the version string (e.g., v1.0.0-5-g123456)
# This makes your builds reproducible and version-aware.
VERSION ?= $(shell git describe --tags --always --dirty)
# Get the current date for the build time
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# --- Go Build Flags ---
# LDFLAGS are linker flags. The -X flag allows us to inject variable values into our Go program at build time.
# We will inject the version and build date. Your Go code needs to have corresponding variables.
LDFLAGS = -ldflags="-X 'main.version=$(VERSION)' -X 'main.buildDate=$(BUILD_DATE)'"

# --- Targets ---

# The .PHONY directive tells make that these targets do not produce files with the same name.
.PHONY: all build-linux build-mac clean help

# The default target, executed when you just run `make`. It builds for all specified platforms.
all: build-linux build-mac ## Build for all target platforms (Linux amd64, macOS arm64)

# Build for Linux amd64
build-linux: ## Build for Linux (amd64)
	@echo "==> Building for Linux (amd64)..."
	@# CGO_ENABLED=0 ensures a statically linked binary without C dependencies.
	@# GOOS and GOARCH specify the target operating system and architecture.
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

# Build for macOS arm64 (Apple Silicon)
build-mac: ## Build for macOS (arm64)
	@echo "==> Building for macOS (arm64)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-mac-arm64 .

# Clean up build artifacts
clean: ## Clean up the build directory
	@echo "==> Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Self-documenting help target
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'