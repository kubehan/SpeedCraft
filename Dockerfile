# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

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
    go build -trimpath \
      -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
      -o /speedcraft .

FROM alpine:3.19
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"

RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -H -h /app speedcraft

WORKDIR /app
COPY --from=builder /speedcraft .
COPY templates ./templates
COPY static ./static
RUN mkdir -p data static/uploads && chown -R speedcraft:speedcraft /app

USER speedcraft
EXPOSE 8080

ENV PORT=8080 \
    DB_PATH=/app/data/speedcraft.db \
    APP_VERSION="${VERSION}" \
    APP_BUILD_DATE="${BUILD_DATE}" \
    APP_GIT_COMMIT="${GIT_COMMIT}"

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/ >/dev/null 2>&1 || exit 1

CMD ["./speedcraft"]
