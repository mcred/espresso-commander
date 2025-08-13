GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

.PHONY: build test

build:
	$(GOBUILD) -o $(PWD)/bin/esp-commander .

test:
	$(GOTEST) -v ./...
