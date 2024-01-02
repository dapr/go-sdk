package database

import (
	"context"
	"fmt"
	"log"
	"os"

	//"github.com/jackc/pgx"

	"github.com/jackc/pgx/v5/pgxpool"
	//_ "github.com/lib/pq"
	//_ "github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5"
	//_ "github.com/jackc/pgx/v5/stdlib"
)

func OpenDBConnection(connectionString string) *pgxpool.Pool {

	fmt.Printf("Database URL: %s\n", os.Getenv("DATABASE_URL"))
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	err = dbpool.Ping(context.Background())

	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return dbpool
}
