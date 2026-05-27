COVERAGE_EXCLUDE   = mocks|main.go|test|pkg/
COVERAGE_THRESHOLD = 80

# ─── Docker image config ──────────────────────────────────────────
# DOCKER_USER : Docker Hub username (override with `make docker-build DOCKER_USER=foo`)
# IMAGE_NAME  : repo name on Docker Hub
# GIT_SHA     : short commit SHA — used to dual-tag image (rollback-able)
DOCKER_USER  ?= jaimeshub
IMAGE_NAME   ?= bookmark-app
DOCKER_IMAGE := bookmark-api
DOCKER_TAG   := latest
GIT_SHA      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
IMAGE_SHA    := $(DOCKER_USER)/$(IMAGE_NAME):$(GIT_SHA)
IMAGE_LATEST := $(DOCKER_USER)/$(IMAGE_NAME):latest

.PHONY: help run test test-race test-service test-repository test-handler test-coverage build clean install-tools swag fmt generate \
        docker-compose-build docker-run docker-stop docker-logs docker-ps \
        docker-build docker-push docker-buildx-multiarch-push docker-run-local docker-clean

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
	@echo ""
	@echo "  Docker (compose, local dev):"
	@echo "  make docker-compose-build - Build app image via compose"
	@echo "  make docker-run           - Start app + Redis via compose (requires .env)"
	@echo "  make docker-stop          - Stop and remove compose services"
	@echo "  make docker-logs          - Tail app container logs"
	@echo "  make docker-ps            - List running compose services"
	@echo ""
	@echo "  Docker (Hub publish — dual-tag SHA + latest):"
	@echo "  make docker-build                 - Build image (host arch) $(IMAGE_SHA) + $(IMAGE_LATEST)"
	@echo "  make docker-push                  - Push both tags to Docker Hub (run docker login first)"
	@echo "  make docker-buildx-multiarch-push - ⭐ Build amd64+arm64 multi-arch + push (RECOMMENDED khi dev Mac M-series)"
	@echo "  make docker-run-local             - Run latest tag against Redis on host (smoke test)"
	@echo "  make docker-clean                 - Remove local image tags"

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
	go test ./... -coverprofile=coverage.tmp -covermode=atomic -coverpkg=./... -p 1
	grep -vE "$(COVERAGE_EXCLUDE)" coverage.tmp > coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@total=$$(go tool cover -func=coverage.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage ($$total%) is below threshold ($(COVERAGE_THRESHOLD)%)"; \
		exit 1; \
	else \
		echo "✅ Coverage ($$total%) meets threshold ($(COVERAGE_THRESHOLD)%)"; \
	fi

generate:
	go generate ./...

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	@echo "Build output: bin/api"

clean:
	rm -rf bin/ coverage.tmp coverage.out coverage.html

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

# ─── Docker targets (compose, local dev) ────────────────────────
docker-compose-build:
	docker compose build
	@echo "✅ Image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run:
	@[ -f .env ] || { echo "❌ .env not found. Run: cp .env.example .env"; exit 1; }
	docker compose up -d
	@. ./.env && echo "✅ Services started. App: http://localhost:$$HOST_PORT | Logs: make docker-logs"

docker-stop:
	docker compose down
	@echo "✅ Services stopped"

docker-logs:
	docker compose logs -f app

docker-ps:
	docker compose ps

# ─── Docker targets (Hub publish — dual-tag SHA + latest) ───────
# `:latest` alone is an anti-pattern (no rollback, no traceability).
# Tag with both `<git-sha>` and `latest` so we can pin VM deploys to a SHA
# and still let "latest" point at the newest known-good build.
docker-build:
	@echo ">> Building $(IMAGE_SHA) + $(IMAGE_LATEST)"
	docker build \
	  -t $(IMAGE_SHA) \
	  -t $(IMAGE_LATEST) \
	  .
	@echo ">> Image size:"
	@docker images $(IMAGE_SHA) --format "  {{.Repository}}:{{.Tag}} = {{.Size}}"

docker-push:
	@echo ">> Verify Docker Hub login..."
# 	@docker info 2>/dev/null | grep -q "Username" || { echo "❌ Not logged in. Run: docker login"; exit 1; }
	@cat ~/.docker/config.json 2>/dev/null | grep -q '"https://index.docker.io/v1/"' || \
    { echo "❌ Not logged in to Docker Hub. Run: docker login"; exit 1; }
	@echo ">> Pushing $(IMAGE_SHA)"
	docker push $(IMAGE_SHA)
	@echo ">> Pushing $(IMAGE_LATEST)"
	docker push $(IMAGE_LATEST)
	@echo "✅ Pushed. Verify: https://hub.docker.com/r/$(DOCKER_USER)/$(IMAGE_NAME)/tags"

# ─── Docker buildx (multi-arch publish) ──────────────────────────
# Build amd64 + arm64 cùng lúc + push trực tiếp lên Hub trong 1 lệnh.
# Image chạy native cả Mac arm64 (dev local) + VM amd64 (prod) — không qua qemu emulation.
# RECOMMENDED khi `bookmark-deployment/docker-compose.yml` KHÔNG có `platform:` pin
# (Docker tự pick native arch từ manifest list).
# Trade-off: build chậm hơn ~1.5–2x vs single-arch (build 2 manifest), nhưng one-time cost
# đổi lấy dev experience nhanh hơn về sau (no qemu).
# Ref: assignments/Lecture-04-docker-rebuild-publish.md § Phase 3c (Option C)
docker-buildx-multiarch-push:
	@echo ">> Verify Docker Hub login..."
	@cat ~/.docker/config.json 2>/dev/null | grep -q '"https://index.docker.io/v1/"' || \
    { echo "❌ Not logged in to Docker Hub. Run: docker login"; exit 1; }
	@echo ">> Verify buildx builder available..."
	@docker buildx inspect default >/dev/null 2>&1 || \
    { echo "❌ buildx default builder not found. Run: docker buildx create --use --name multiarch"; exit 1; }
	@echo ">> Building amd64+arm64 multi-arch + push $(IMAGE_SHA) + $(IMAGE_LATEST)"
	docker buildx build \
	  --platform linux/amd64,linux/arm64 \
	  -t $(IMAGE_SHA) \
	  -t $(IMAGE_LATEST) \
	  --push \
	  .
	@echo "✅ Pushed multi-arch image. Verify with:"
	@echo "   docker buildx imagetools inspect $(IMAGE_SHA) | grep Platform:"
	@echo "   Kỳ vọng 2 dòng: linux/amd64 + linux/arm64 (bỏ qua unknown/unknown — SLSA attestation)"

docker-run-local:
	docker run --rm -it \
	  -e REDIS_ADDR=host.docker.internal:6379 \
	  -e SERVICE_NAME=bookmark-local \
	  -e APP_HOSTNAME=docker-local-test \
	  -p 8080:8080 \
	  $(IMAGE_LATEST)

docker-clean:
	-docker rmi $(IMAGE_SHA) $(IMAGE_LATEST) 2>/dev/null || true
