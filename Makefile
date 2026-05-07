.PHONY: help run test test-coverage build clean install-tools swag fmt

help:
	@echo "Available targets:"
	@echo "  make run                  - Run the API server (regenerates Swagger docs)"
	@echo "  make test                 - Run all tests"
	@echo "  make test-coverage        - Run tests with coverage report"
	@echo "  make build                - Build the binary"
	@echo "  make clean                - Remove build artifacts and coverage"
	@echo "  make swag                 - Generate Swagger documentation"
	@echo "  make install-tools        - Install development tools"
	@echo "  make fmt                  - Format code"

run: swag
	go run ./cmd/api

test:
	go test -v ./...

test-coverage:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	@echo "Build output: bin/api"

clean:
	rm -rf bin/ coverage.out coverage.html

swag:
	swag init -g cmd/api/main.go

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/vektra/mockery/v2@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development tools installed successfully"

fmt:
	go fmt ./...
	goimports -w .
