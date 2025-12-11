# Makefile for agentlog project

.PHONY: all build test clean install help

# Default target
all: build

# Build the agentlog binary
build:
	@echo "Building agentlog..."
	go build -o agentlog ./cmd/agentlog

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Install agentlog to GOPATH/bin
install: build
	@echo "Installing agentlog to $$(go env GOPATH)/bin..."
	go install ./cmd/agentlog

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f agentlog

# Show help
help:
	@echo "Agentlog Makefile targets:"
	@echo "  make build   - Build the agentlog binary"
	@echo "  make test    - Run all tests"
	@echo "  make install - Install agentlog to GOPATH/bin"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make help    - Show this help message"
