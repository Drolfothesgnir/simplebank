package main

import (
	"database/sql"
	"log"

	"github.com/Drolfothesgnir/simplebank/api"
	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to the database: ", err)
	}

	db := db.NewStore(conn)
	server := api.NewServer(db)
	if err := server.Start(serverAddress); err != nil {
		log.Fatal("Cannost start the server:", err)
	}
}
