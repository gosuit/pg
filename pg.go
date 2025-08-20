package pg

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Query(sql string, model any) Query
	Command(sql string, model any) Command
	Transactional(ctx context.Context, fn TxFunc) error

	ToPgx() *pgxpool.Pool
	ToDB() *sql.DB
}

func New() Client {
	return &client{}
}

type Query interface {
	WithArgs(args ...*Argument) Query
	WithArg(key string, value any) Query
	Exec(ctx context.Context) error
}

type Command interface {
	WithArgs(args ...*Argument) Command
	WithArg(key string, value any) Command
	Returning(dest any) Command
	Exec(ctx context.Context) error
}

func Arg(key string, value any) *Argument {
	return &Argument{
		key,
		value,
	}
}

type Argument struct {
	key   string
	value any
}

type TxFunc func(ctx context.Context) error
