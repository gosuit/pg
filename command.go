package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Command interface {
	WithArgs(args ...*Argument) Command
	WithArg(key string, value any) Command
	Returning(dest any) Command
	Exec(ctx context.Context) error
}

type command struct {
	pool  *pgxpool.Pool
	sql   string
	model any
	dest  any
	args  map[string]any
}

func (c *command) WithArgs(args ...*Argument) Command {
	for _, a := range args {
		c.args[a.key] = a.value
	}

	return c
}

func (c *command) WithArg(key string, value any) Command {
	c.args[key] = value

	return c
}

func (c *command) Returning(dest any) Command {
	c.dest = dest

	return c
}

func (c *command) Exec(ctx context.Context) error {
	return nil
}
