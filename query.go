package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Query interface {
	WithArgs(args ...*Argument) Query
	WithArg(key string, value any) Query
	Exec(ctx context.Context) error
}

type query struct {
	pool  *pgxpool.Pool
	sql   string
	model any
	args  map[string]any
}

func (q *query) WithArgs(args ...*Argument) Query {
	for _, a := range args {
		q.args[a.key] = a.value
	}

	return q
}

func (q *query) WithArg(key string, value any) Query {
	q.args[key] = value

	return q
}

func (q *query) Exec(ctx context.Context) error {
	q.getQueryManager(ctx)

	return nil
}

func (q *query) getQueryManager(ctx context.Context) queryManager {
	tx, ok := getTxFromContext(ctx)
	if ok {
		return tx
	}

	return q.pool
}
