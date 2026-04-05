.PHONY: build test run up down vet

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/petstore

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -timeout 120s

run: up
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/server

up:
	docker compose up -d

down:
	docker compose down

vet:
	go vet ./...
