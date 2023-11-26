package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var connPool *pgxpool.Pool

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("fatal error while reading config file", err)
	}

	connPool, err = pgxpool.New(context.Background(), config.DbUrl)
	defer connPool.Close()
	if err != nil {
		log.Fatal("Failed to connect to db", err)
	}

	testQueries = New(connPool)

	os.Exit(m.Run())
}
