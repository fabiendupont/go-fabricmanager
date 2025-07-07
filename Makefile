# NVIDIA FabricManager Go Package Makefile

.PHONY: all build clean test examples install

# Build configuration
BINARY_NAME=fmpm
PACKAGE_NAME=github.com/NVIDIA/go-fabricmanager
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
CGO_ENABLED=1

# Default target
all: build

# Build the CLI tool
build: $(BINARY_NAME)

$(BINARY_NAME): fabricmanager.go cmd/fmpm/main.go
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BINARY_NAME) cmd/fmpm/main.go

# Build examples
examples: examples/basic_usage/$(BINARY_NAME) examples/error_handling/$(BINARY_NAME)

examples/basic_usage/$(BINARY_NAME): examples/basic_usage/main.go fabricmanager.go
	@echo "Building basic usage example..."
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $@ $<

examples/error_handling/$(BINARY_NAME): examples/error_handling/main.go fabricmanager.go
	@echo "Building error handling example..."
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $@ $<

# Run tests
test:
	@echo "Running tests..."
	CGO_ENABLED=$(CGO_ENABLED) go test -v ./...

# Install the CLI tool
install: $(BINARY_NAME)
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f examples/basic_usage/$(BINARY_NAME)
	rm -f examples/error_handling/$(BINARY_NAME)
	go clean -cache

# Run the CLI tool
run: build
	./$(BINARY_NAME) --help

# Run examples
run-basic: examples/basic_usage/$(BINARY_NAME)
	./examples/basic_usage/$(BINARY_NAME)

run-error-handling: examples/error_handling/$(BINARY_NAME)
	./examples/error_handling/$(BINARY_NAME)

# Development helpers
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Check if FabricManager development package is installed
check-deps:
	@echo "Checking dependencies..."
	@if [ ! -f "/usr/include/nv_fm_agent.h" ]; then \
		echo "Error: NVIDIA FabricManager development package not found."; \
		echo "Please install nvidia-fabricmanager-devel package."; \
		exit 1; \
	fi
	@if [ ! -f "/usr/lib/x86_64-linux-gnu/libnvfm.so" ] && [ ! -f "/usr/lib/aarch64-linux-gnu/libnvfm.so" ]; then \
		echo "Error: NVIDIA FabricManager library not found."; \
		echo "Please install nvidia-fabricmanager package."; \
		exit 1; \
	fi
	@echo "Dependencies check passed."

# Build with debug information
debug: LDFLAGS += -gcflags="all=-N -l"
debug: build

# Build for different architectures
build-linux-amd64:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 cmd/fmpm/main.go

build-linux-arm64:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 cmd/fmpm/main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the CLI tool"
	@echo "  examples           - Build example programs"
	@echo "  test               - Run tests"
	@echo "  install            - Install the CLI tool"
	@echo "  clean              - Clean build artifacts"
	@echo "  run                - Run the CLI tool with help"
	@echo "  run-basic          - Run basic usage example"
	@echo "  run-error-handling - Run error handling example"
	@echo "  deps               - Install Go dependencies"
	@echo "  check-deps         - Check system dependencies"
	@echo "  debug              - Build with debug information"
	@echo "  build-linux-amd64  - Build for Linux AMD64"
	@echo "  build-linux-arm64  - Build for Linux ARM64"
	@echo "  help               - Show this help message" 