package main

import (
	"context"
	"log"

	"github.com/gosuit/pg"
)

type User struct {
	Name     string `pg:"name"`
	Password string `pg:"password"`
}

func main() {
	ctx := context.Background()

	cfg := &pg.Config{
		Host:     "localhost",
		Port:     5432,
		DBName:   "postgres",
		Username: "admin",
		Password: "root",
		SSLMode:  "disable",
	}

	// Init client
	client, err := pg.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// You can set values to query with "@" prefix.
	// Also you can set values with args like in query
	sql := "INSERT INTO users VALUES (@name, @password)"
	u := User{
		Name:     "admin",
		Password: "root",
	}

	err = client.Command(sql, &u).Exec(ctx)
	if err != nil {
		panic(err)
	}

	// If sql-query has RETURNING command, you can set destination with pg.Command.Returning
}
