.PHONY: help run test test-race test-service test-repository test-handler test-coverage build clean install-tools swag fmt generate

help:
	@echo "Available targets:"
	@echo "  make run                  - Run the API server (regenerates Swagger docs)"
	@echo "  make test                 - Run all tests"
	@echo "  make test-race            - Run all tests with race detection"
	@echo "  make test-service         - Run service layer tests (verbose + race + cover)"
	@echo "  make test-repository      - Run repository layer tests (verbose + race + cover)"
	@echo "  make test-handler         - Run handler layer tests (verbose + race + cover)"
	@echo "  make test-coverage        - Run tests with HTML coverage report"
	@echo "  make build                - Build the binary"
	@echo "  make clean                - Remove build artifacts and coverage"
	@echo "  make swag                 - Generate Swagger documentation"
	@echo "  make generate             - Run go generate (regenerate mocks, etc.)"
	@echo "  make install-tools        - Install development tools"
	@echo "  make fmt                  - Format code"

run: swag
	go run ./cmd/api

test:
	go test -v ./...

test-race:
	go test -v -race ./...

test-service:
	go test ./internal/service/... -v -race -cover

test-repository:
	go test ./internal/repository/... -v -race -cover

test-handler:
	go test ./internal/handler/... -v -race -cover

test-coverage:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

generate:
	go generate ./...

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
