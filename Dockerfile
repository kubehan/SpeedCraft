FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /speedcraft .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -H -h /app speedcraft
WORKDIR /app
COPY --from=builder /speedcraft .
COPY templates ./templates
COPY static ./static
RUN mkdir -p data && chown -R speedcraft:speedcraft /app

USER speedcraft
EXPOSE 8080

ENV PORT=8080
ENV DB_PATH=/app/data/speedcraft.db

CMD ["./speedcraft"]
