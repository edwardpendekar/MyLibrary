# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS builder
WORKDIR /src

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata python3 py3-pip \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api ./api
COPY migration ./migration
COPY scripts ./scripts

# The admin "translate a book from a document" feature shells out to
# scripts/translate_book.py (docx parsing + a plain REST call to the Gemini
# API — no heavy SDK needed). A venv keeps pip off the system Python that
# Alpine itself manages.
RUN python3 -m venv /opt/venv \
    && /opt/venv/bin/pip install --no-cache-dir -r /app/scripts/requirements.txt
ENV PATH="/opt/venv/bin:$PATH"

RUN mkdir -p /app/storage/uploads && chown -R app:app /app
USER app

EXPOSE 8080
ENTRYPOINT ["/app/api"]
