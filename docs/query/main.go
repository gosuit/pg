package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gosuit/pg/v2"
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

	// You can set values with model values with "@" prefix as in command
	// Also you can set values with args with "#" prefix
	sql := "SELECT * FROM users WHERE name = #name"
	var u User

	// This will be set fields of model with values from result from DB.
	// For this, value of "pg" tag must be the same with column name (or column alias) from DB.
	err = client.Query(sql, &u).WithArg("name", "admin").Exec(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(u)
}
