GOCMD=go
GOTEST=$(GOCMD) test

.PHONY: test

test:
	$(GOTEST) -v ./...
