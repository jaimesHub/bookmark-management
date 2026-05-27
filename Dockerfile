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

# ca-certificates: HTTPS outbound calls
# wget: needed by docker-compose / VM healthcheck probes (Task #2)
RUN apk add --no-cache ca-certificates wget && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api /usr/local/bin/api

USER app

# Port (EXPOSE) and HEALTHCHECK are intentionally configured at the
# orchestration layer (docker-compose.yml) rather than baked into the
# image — keeps the image port-agnostic and lets compose interpolate
# ${HOST_PORT}/${API_CONTAINER_PORT} from .env at runtime.

ENTRYPOINT ["/usr/local/bin/api"]
