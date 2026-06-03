# syntax=docker/dockerfile:1.7

# ---- Stage 1: builder ----
# Pin Go 1.26.2 to match local asdf .tool-versions + course spec.
# Bump manually when intentionally upgrading; floating `golang:1.26-alpine`
# would drift to latest 1.26.x silently (reproducibility risk).
FROM golang:1.26.2-alpine AS builder

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

# ca-certificates: HTTPS outbound calls
# wget: needed by docker-compose / VM healthcheck probes (Task #2)
RUN apk add --no-cache ca-certificates wget && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api /usr/local/bin/api

# Lec-6: WORKDIR /app áp dụng cho runtime stage để file mount tại
# /app/keys/* (RSA private/public key) resolve đúng khi env path dùng dạng
# relative `./keys/...`. Builder stage có WORKDIR /app riêng — KHÔNG carry
# sang runtime stage tự động khi `FROM alpine:3.20` reset state.
WORKDIR /app

USER app

# HEALTHCHECK is intentionally configured at the orchestration layer
# (docker-compose.yml) rather than baked into the image — keeps the image
# port-agnostic and lets compose interpolate ${HOST_PORT}/${API_CONTAINER_PORT}
# from .env at runtime.
#
# EXPOSE 8080 is documentary metadata only — it does NOT publish the port
# to host. Port accessibility is controlled by compose `ports:` (currently
# commented out for the app service, so :8080 stays internal to the docker
# network — only nginx :80 is host-mapped) + VM firewall (only 22/80 open).
# Listed here so `docker inspect`, Kubernetes, and onboarding devs can
# auto-detect which port the service listens on.
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/api"]
