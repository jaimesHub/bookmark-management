# syntax=docker/dockerfile:1.7

# ---- Stage 1: builder ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache go mod download as a separate layer:
# only re-runs when go.mod / go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Build with production flags:
#   - CGO_ENABLED=0 → static binary, alpine-compatible
#   - -ldflags="-s -w" → strip debug symbols (~30% smaller)
#   - -trimpath → reproducible build, strip absolute paths
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /out/api \
    ./cmd/api

# ---- Stage 2: runtime ----
FROM alpine:3.20

# Non-root user:
# container compromise should not give attacker root.
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api /usr/local/bin/api

USER app

EXPOSE 8080

# HEALTHCHECK — reuse /health-check endpoint from Lec 1.
# wget is built into alpine (busybox).
# Compose `depends_on.condition: service_healthy` relies on this directive.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health-check || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
