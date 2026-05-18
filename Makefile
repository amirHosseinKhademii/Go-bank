ifneq (,$(wildcard .env))
include .env
export
endif

postgres:
	docker run --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=admin -p 5432:5432 -d postgres

createdb:
	docker exec -it postgres createdb --username=root --owner=root bank

dropdb:
	docker exec -it postgres dropdb --username=root --owner=root bank

migrateup:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up

migrateup1:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up 1

migratedown:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down

migratedown1:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down 1

migrateTestDbup:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ -z "$(DB_TEST)" ]; then \
		echo "Error: DB_TEST is not set"; \
		exit 1; \
	fi
	migrate -path db/migration -database "$(DB_TEST)" -verbose up

migrateTestDbdown:
	@if ! command -v migrate 2>/dev/null; then \
		echo "Installing migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ -z "$(DB_TEST)" ]; then \
		echo "Error: DB_TEST is not set"; \
		exit 1; \
	fi
	migrate -path db/migration -database "$(DB_TEST)" -verbose down -all

sqlc:
	sqlc generate

test:
	go test -v ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./...

dev:
	go run ./cmd/main.go

build:
	go build -o bin/app ./cmd/main.go

validate-infra:
	./scripts/validate-infrastructure.sh

.PHONY: postgres createdb dropdb migrateup1 migratedown1 test lint sqlc dev build migrateTestDbup migrateTestDbdown validate-infra