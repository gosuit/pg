package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
	return nil
}
