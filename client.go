package pg

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type client struct {
	pool *pgxpool.Pool
}

func (c *client) Query(sql string, model any) Query {
	return &query{
		pool:  c.pool,
		sql:   sql,
		model: model,
	}
}

func (c *client) Command(sql string, model any) Command {
	return &command{
		pool:  c.pool,
		sql:   sql,
		model: model,
	}
}

func (c *client) Transactional(ctx context.Context, fn TxFunc) error {
	return nil
}

func (c *client) ToPgx() *pgxpool.Pool {
	return c.pool
}

func (c *client) ToDB() *sql.DB {
	return stdlib.OpenDBFromPool(c.pool)
}
