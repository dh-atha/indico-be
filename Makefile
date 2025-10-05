SHELL := /bin/bash

# Config
APP_NAME ?= be
PORT ?= 8080
WORKERS ?= 8
DATABASE_URL ?= postgres://postgres:password@localhost:5432/appdb?sslmode=disable

.PHONY: help
help:
	@echo "Targets:"
	@echo "  db-up           - start postgres via docker compose"
	@echo "  migrate         - apply all SQL migrations in order"
	@echo "  run             - run the server locally"
	@echo "  build           - build local binary to bin/server"
	@echo "  docker-build    - build docker image ($(APP_NAME):latest)"
	@echo "  docker-run      - run container with host network (Linux)"
	@echo "  docker-stop     - stop running container"

.PHONY: db-up
db-up:
	docker compose up -d db

.PHONY: migrate
migrate:
	@if ! command -v psql >/dev/null 2>&1; then echo "psql not found. Please install PostgreSQL client."; exit 1; fi
	@set -euo pipefail; \
	for f in $(shell ls -1 migrations/*.sql | sort); do \
		echo "Applying $$f"; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f "$$f"; \
	done

.PHONY: run
run:
	PORT=$(PORT) WORKERS=$(WORKERS) DATABASE_URL=$(DATABASE_URL) go run ./...

.PHONY: build
build:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/server ./

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run:
	# On Linux, --network host lets the app reach localhost:5432
	docker run --rm -e PORT=$(PORT) -e WORKERS=$(WORKERS) -e DATABASE_URL=$(DATABASE_URL) --network host -p $(PORT):$(PORT) -v $$PWD/tmp:/app/tmp --name $(APP_NAME) $(APP_NAME):latest

.PHONY: docker-stop
docker-stop:
	-@docker rm -f $(APP_NAME) >/dev/null 2>&1 || true
