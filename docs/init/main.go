package main

import (
	"context"
	"log"

	"github.com/gosuit/pg"
)

func main() {
	ctx := context.Background()

	// pg.Config support "confy" and other configuration tags.
	cfg := &pg.Config{
		Host:     "localhost",
		Port:     5432,
		DBName:   "your_database",
		Username: "your_username",
		Password: "your_password",
		SSLMode:  "disable",
	}

	// Create client.
	client, err := pg.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
 
	// Use or convert client.
	client.ToPgx() // *pgxpool.Pool will be returned.
	client.ToDB()  // *sql.DB will be returned.
}
