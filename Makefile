GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=espresso-commander
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: build test clean install uninstall package run

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