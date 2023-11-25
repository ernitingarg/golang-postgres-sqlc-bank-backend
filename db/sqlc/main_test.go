package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var connPool *pgxpool.Pool

const dataSourceName = "postgresql://admin:password123@localhost:5432/postgresdb?sslmode=disable"

func TestMain(m *testing.M) {
	var err error
	connPool, err = pgxpool.New(context.Background(), dataSourceName)
	defer connPool.Close()
	if err != nil {
		log.Fatal("Failed to connect to db", err)
	}

	testQueries = New(connPool)

	os.Exit(m.Run())
}
