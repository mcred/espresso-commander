GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=espresso-commander
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: build test clean install uninstall package run help

# Default target
all: test build

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p bin
	$(GOBUILD) -o $(PWD)/bin/$(BINARY_NAME) -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" .

test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -cover ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@$(GOCMD) clean

run: build
	./bin/$(BINARY_NAME)

# macOS installation targets
install: build
	@echo "Installing $(BINARY_NAME)..."
	@chmod +x installer/install.sh
	@sudo ./installer/install.sh

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@chmod +x installer/uninstall.sh
	@sudo ./installer/uninstall.sh

# Create macOS .pkg installer
package: build
	@echo "Creating macOS package..."
	@chmod +x installer/build-pkg.sh
	@./installer/build-pkg.sh

# Development helpers
dev: build
	@echo "Running in development mode..."
	./bin/$(BINARY_NAME)

logs:
	@echo "Showing service logs..."
	@tail -f /var/log/espresso-commander.log

status:
	@echo "Checking service status..."
	@sudo launchctl list | grep io.mcred || echo "Service not running"

help:
	@echo "Espresso Commander - Makefile targets"
	@echo ""
	@echo "Building:"
	@echo "  make build         - Build the binary"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make all          - Run tests and build"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo ""
	@echo "Running:"
	@echo "  make run          - Build and run"
	@echo "  make dev          - Run in development mode"
	@echo ""
	@echo "Installation (macOS):"
	@echo "  make install      - Install as system service (requires sudo)"
	@echo "  make uninstall    - Uninstall system service (requires sudo)"
	@echo "  make package      - Create .pkg installer for distribution"
	@echo ""
	@echo "Service Management:"
	@echo "  make status       - Check service status"
	@echo "  make logs         - Tail service logs"
	@echo ""
	@echo "Other:"
	@echo "  make help         - Show this help message"
