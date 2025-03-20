package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB(connString string) error {
	var err error
	Pool, err = pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		return err
	}
	log.Println("Connected to messenger database")
	return nil
}
