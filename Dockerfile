# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build for target platform
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /speedcraft .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -H -h /app speedcraft

WORKDIR /app
COPY --from=builder /speedcraft .
COPY templates ./templates
COPY static ./static
RUN mkdir -p data && chown -R speedcraft:speedcraft /app

USER speedcraft
EXPOSE 8080

ENV PORT=8080 \
    DB_PATH=/app/data/speedcraft.db

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/ >/dev/null 2>&1 || exit 1

CMD ["./speedcraft"]
