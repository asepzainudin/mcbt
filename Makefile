.PHONY: help run build test lint fmt vet tidy migrate-up migrate-down migrate-status migrate-new db-drop-recreate

APP_NAME     := mcbt
BUILD_DIR    := bin
MIGRATE_CMD  := go run ./cmd/migrate
NEXT_SEQ     := $(shell ls migrations | sort -n | tail -1 | cut -d'_' -f1 | awk '{printf "%06d", $$1+1}')

help:
	@echo "Available targets:"
	@echo "  make run              - start API server"
	@echo "  make build            - compile binary to ./bin/api"
	@echo "  make test             - run all tests"
	@echo "  make lint             - gofmt check + go vet"
	@echo "  make fmt              - format all code"
	@echo "  make tidy             - clean up module dependencies"
	@echo "  make migrate-up       - apply all pending migrations"
	@echo "  make migrate-down     - revert last migration (steps=N for more)"
	@echo "  make migrate-status   - show current migration version"
	@echo "  make migrate-new NAME - create empty up/down migration files"

run:
	$(info Starting $(APP_NAME) API server...)
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/api ./cmd/api

test:
	go test ./... -v

lint:
	@out=$$(gofmt -l ./cmd ./internal); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

migrate-up:
	$(MIGRATE_CMD) -command up

migrate-down:
	$(MIGRATE_CMD) -command down -steps $(or $(steps),1)

migrate-status:
	$(MIGRATE_CMD) -command status

migrate-new:
	@test -n "$(NAME)" || { echo "usage: make migrate-new NAME=create_users"; exit 1; }
	@touch migrations/$(NEXT_SEQ)_$(NAME).up.sql migrations/$(NEXT_SEQ)_$(NAME).down.sql
	@echo "created migrations/$(NEXT_SEQ)_$(NAME).up.sql and .down.sql"

db-drop-recreate:
	@docker exec postgres psql -U root -c "DROP DATABASE IF EXISTS $(APP_NAME);" \
		&& docker exec postgres psql -U root -c "CREATE DATABASE $(APP_NAME);" \
		&& echo "database recreated"
