# PG

pg is a wrapper around the <a href="https://github.com/jackc/pgx">pgx</a> library for PostgreSQL in Go. It simplifies the process of connecting to a PostgreSQL database, managing migrations, and other. This library provides a clean interface for database operations while leveraging the performance and features of pgx.

## Installation

```zsh
go get github.com/gosuit/pg
```

## Features

• Connection Management: Easily establish connections to PostgreSQL databases using configurable parameters.

• Migration Support: Automatically run database migrations using the goose migration tool.

• Other: Register custom PostgreSQL types with your database connection.

• Mocking: You can create mock client with mock pgx pool

## Usage

```golang
package main

import (
    "context"
    "log"

    "github.com/gosuit/pg"
)

func main() {
    ctx := context.Background()

    cfg := &pg.Config{
        Host:           "localhost",
        Port:           5432,
        DBName:         "your_database",
        Username:       "your_username",
        Password:       "your_password",
        SSLMode:        "disable",
        MigrationsRun:  true,
        MigrationsPath: "./migrations",
    }

    client, err := pg.New(ctx, cfg)
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }

    err = client.RegisterTypes([]string{"custom_type"})
    if err != nil {
        log.Fatalf("failed to register types: %v", err)
    }

    // Use client for database operations...

    // Access underlying pgxpool
    pool := client.ToPgx()
    // Use pool...
}
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
