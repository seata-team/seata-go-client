.PHONY: help build test clean deps examples examples-all lint fmt

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
	@echo "  examples-all - Run all migrated examples"

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
	@go run examples/*.go basic || echo "Basic example requires running Seata server"
	@echo ""
	@echo "Running migrated HTTP saga..."
	@go run examples/*.go mhttp_saga || true
	@echo "Running migrated gRPC saga..."
	@go run examples/*.go mgrpc_saga || true
	@echo "Running migrated gRPC TCC..."
	@go run examples/*.go mgrpc_tcc || true
	@echo "Running migrated gRPC headers..."
	@go run examples/*.go mgrpc_headers || true
	@echo "Running migrated HTTP msg..."
	@go run examples/*.go mhttp_msg || true
	@echo "Running migrated workflow (gRPC saga)..."
	@go run examples/*.go mgrpc_workflow_saga || true
	@echo "Done."

# Run all migrated examples
examples-all: deps
	@set -e; \
	for ex in \
	  mhttp_saga \
	  mgrpc_saga \
	  mgrpc_tcc \
	  mgrpc_headers \
	  mgrpc_msg \
	  mhttp_headers \
	  mhttp_msg \
	  mgrpc_saga_barrier \
	  mgrpc_saga_other \
	  mhttp_workflow_saga \
	  mhttp_workflow_tcc \
	  mhttp_xa \
	  mhttp_gorm_barrier \
	  mhttp_barrier_redis \
	  mhttp_saga_mongo \
	  mgrpc_workflow_saga \
	  mgrpc_workflow_tcc \
	  mgrpc_workflow_mixed \
	  mhttp_tcc \
	  mhttp_tcc_barrier \
	  mhttp_saga_barrier \
	  mhttp_saga_mutidb \
	  mhttp_xa_gorm \
	  mgrpc_xa \
	  mhttp_more; do \
	  echo "Running $$ex ..."; \
	  go run examples/*.go $$ex || { echo "Example $$ex failed"; exit 1; }; \
	  echo ""; \
	done; \
	echo "All migrated examples completed."

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
