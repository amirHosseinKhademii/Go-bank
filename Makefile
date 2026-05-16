postgres:
	docker run --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=admin -p 5432:5432 -d postgres
createdb:
	docker exec -it postgres createdb --username=root --owner=root bank
dropdb:
	docker exec -it postgres dropdb --username=root --owner=root bank
migrateup:
	migrate -path db/migration -database "postgres://root:admin@localhost:5432/bank?sslmode=disable" -verbose up
.PHONY: postgres createdb dropdb