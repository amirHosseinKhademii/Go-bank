package main

import (
	"bank/api"
	repository "bank/db/sqlc"
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbSource = os.Getenv("DB_SOURCE")

func main() {
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new store
	store := repository.NewStore(conn)

	// Create a new server
	server := api.NewServer(store)

	// Start the server
	err = server.Start(":8080")
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
