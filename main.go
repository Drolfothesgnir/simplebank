package main

import (
	"database/sql"
	"log"

	"github.com/Drolfothesgnir/simplebank/api"
	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot read config file: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the database: ", err)
	}

	db := db.NewStore(conn)
	server := api.NewServer(db)
	if err := server.Start(config.ServerAddress); err != nil {
		log.Fatal("Cannost start the server:", err)
	}
}
