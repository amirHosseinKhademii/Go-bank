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
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up
migratedown:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down
migrateTestDbup:
	migrate -path db/migration -database "$(DB_TEST)" -verbose up
migrateTestDbdown:
	migrate -path db/migration -database "$(DB_TEST)" -verbose down -all
sqlc:
	sqlc generate
test:
	go test -v ./... -cover
lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./...
dev:
	go run ./cmd/main.go
build:
	go build -o bin/app ./cmd/main.go
.PHONY: postgres createdb dropdb migrateup migratedown test lint sqlc dev build migrateTestDbup migrateTestDbdown