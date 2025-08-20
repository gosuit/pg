package pg

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type client struct {
	pool *pgxpool.Pool
}

func (c *client) ToPgx() *pgxpool.Pool {
	return c.pool
}

func (c *client) ToDB() *sql.DB {
	return stdlib.OpenDBFromPool(c.pool)
}
