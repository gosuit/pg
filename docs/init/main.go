package main

import (
	"context"
	"log"

	"github.com/gosuit/pg"
)

func main() {
	ctx := context.Background()

	cfg := &pg.Config{
		Host:     "localhost",
		Port:     5432,
		DBName:   "your_database",
		Username: "your_username",
		Password: "your_password",
		SSLMode:  "disable",
	}

	client, err := pg.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	client.ToPgx()
	client.ToDB()
}
