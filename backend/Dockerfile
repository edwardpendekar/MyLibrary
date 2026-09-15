# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS builder
WORKDIR /src

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api ./api
COPY migration ./migration

RUN mkdir -p /app/storage/uploads && chown -R app:app /app
USER app

EXPOSE 8080
ENTRYPOINT ["/app/api"]
