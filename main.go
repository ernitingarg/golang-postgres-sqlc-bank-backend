package main

import (
	"context"
	"log"

	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/api"
	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := utils.LoadConfig("./")
	if err != nil {
		log.Fatal("fatal error while reading config file", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DbUrl)
	defer connPool.Close()
	if err != nil {
		log.Fatal("Failed to connect to db", err)
	}

	store := db.NewStore(connPool)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Failed to start the server", err)
	}
}
