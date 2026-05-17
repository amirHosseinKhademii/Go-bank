# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All workflows are driven by the [Makefile](Makefile). The `.env` file is auto-loaded and exported by `make`, so `DB_SOURCE` is available to targets.

- `make postgres` — start a local Postgres 5432 container (user `root`, password `admin`)
- `make createdb` / `make dropdb` — create/drop the `bank` database inside the container
- `make migrateup` / `make migratedown` — run `golang-migrate` against `db/migration` using `$DB_SOURCE`
- `make sqlc` — regenerate Go code from SQL (driven by [sqlc.yaml](sqlc.yaml))
- `make test` — `go test -v ./...` (requires a reachable DB at `$DB_SOURCE`)
- Run a single test: `go test -v ./db/sqlc -run TestCreateAccount`

`DB_SOURCE` must be set for both migrations and tests — [db/sqlc/main_test.go](db/sqlc/main_test.go) reads it directly via `os.Getenv` and opens a `pgxpool`. Tests fail with a connect error if `.env` isn't loaded into the shell.

## Architecture

This is the data-access foundation for a bank service (accounts, entries between accounts, transfers between accounts). There is no HTTP/service layer yet — only the DB layer.

**Code generation pipeline.** [sqlc.yaml](sqlc.yaml) wires three directories together:
- [db/migration/](db/migration/) holds the schema (golang-migrate up/down pairs). This is the source of truth for table shape and is also fed to sqlc as `schema:`.
- [db/query/](db/query/) holds hand-written `.sql` files with sqlc annotations (`-- name: Foo :one|:many|:exec`). One file per table.
- [db/sqlc/](db/sqlc/) is the generated Go package `repository` (output of `sqlc generate`). `*.sql.go`, `models.go`, `querier.go`, and `db.go` are generated — **do not hand-edit them**. To change a query or model, edit the SQL in `db/query/` or `db/migration/` and re-run `make sqlc`.

The generated package uses `pgx/v5` (`sql_package: "pgx/v5"`) and emits a `Querier` interface plus a `Queries` struct constructed with `New(DBTX)`. `DBTX` accepts either a `*pgxpool.Pool` or a `pgx.Tx`, and `Queries.WithTx(tx)` returns a transaction-scoped copy — this is the hook for future cross-table transactions (e.g. money transfer = create transfer + two entries + two balance updates).

**Tests.** All `*_test.go` files in [db/sqlc/](db/sqlc/) live in the generated package and exercise real SQL against a real Postgres. `TestMain` opens the pool once; per-test helpers like `createRandomAccount` insert fresh rows. Tests are not hermetic — they leave data behind and assume the schema is migrated up.

**Note on schema.** The `entries` table is misspelled as `entires` in [db/migration/000001_init_schema.up.sql](db/migration/000001_init_schema.up.sql), as is the column `ammount` on `entires` and `transfers`. The Go model names (`Entry`, `Amount`) are correct because they come from the query files — be aware of the mismatch when writing raw SQL or new migrations.
