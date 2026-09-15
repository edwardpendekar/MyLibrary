.PHONY: run build test test-integration lint migrate-up migrate-down migrate-new swagger tidy

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/bookreader?sslmode=disable
MIGRATE := go run -tags postgres -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -short -race -count=1

test-integration:
	go test ./... -run Integration -race -count=1

lint:
	go vet ./...
	gofmt -l .

migrate-up:
	$(MIGRATE) -path migration -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migration -database "$(DATABASE_URL)" down 1

# usage: make migrate-new name=add_something
migrate-new:
	$(MIGRATE) create -ext sql -dir migration -seq $(name)

swagger:
	go run -mod=mod github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs

tidy:
	go mod tidy
