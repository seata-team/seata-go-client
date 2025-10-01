.PHONY: help build test clean deps examples lint fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  deps     - Download dependencies"
	@echo "  build    - Build the project"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linter"
	@echo "  fmt      - Format code"
	@echo "  clean    - Clean build artifacts"
	@echo "  examples - Run examples"

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build the project
build: deps
	go build ./...

# Run tests
test: deps
	go test -v ./...

# Run tests with coverage
test-coverage: deps
	go test -v -cover ./...

# Run linter
lint: deps
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	go clean
	rm -rf dist/

# Run examples
examples: deps
	@echo "Running basic example..."
	@go run examples/main.go basic || echo "Basic example requires running Seata server"
	@echo ""
	@echo "Running saga example..."
	@go run examples/main.go saga || echo "Saga example requires running Seata server"
	@echo ""
	@echo "Running tcc example..."
	@go run examples/main.go tcc || echo "TCC example requires running Seata server"

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks
check: fmt lint test

# Build for different platforms
build-all: deps
	GOOS=linux GOARCH=amd64 go build -o dist/seata-go-client-linux-amd64 ./examples/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/seata-go-client-darwin-amd64 ./examples/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/seata-go-client-windows-amd64.exe ./examples/main.go
