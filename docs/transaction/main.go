package main

import (
	"context"
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
		DBName:   "your_database",
		Username: "your_username",
		Password: "your_password",
		SSLMode:  "disable",
	}

	// Init client
	client, err := pg.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	u1 := User{
		Name:     "admin",
		Password: "root",
	}

	u2 := User{
		Name:     "superAdmin",
		Password: "superRoot",
	}

	// pg.Client.Transactional will wrap the passed func in a transaction.
	//
	// For this you must use the context that the passed func receives in it`s arguments.
	//
	// If necessary, you can setup the transaction using optional arguments:
	//
	//	- WithLevel
	// 	- WithAccess
	//	- WithDeferrable
	// 	- WithBeginQuery
	// 	- WithCommitQuery
	//
	err = client.Transactional(ctx, func(c context.Context) error {
		sql := "INSERT INTO users VALUES (@name, @password)"

		err := client.Command(sql, &u1).Exec(c)
		if err != nil {
			return err
		}

		err = client.Command(sql, &u2).Exec(c)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed tx: %v", err)
	}
}
