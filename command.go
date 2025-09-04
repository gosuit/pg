package pg

import (
	"context"
	"errors"
	"reflect"
)

type Command interface {
	WithArgs(args ...*Argument) Command
	WithArg(key string, value any) Command
	Returning(dest any) Command
	Exec(ctx context.Context) error
}

type command struct {
	client        *client
	sql           string
	src           reflect.Value
	dest          reflect.Value
	withReturning bool
	args          map[string]any
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
	c.dest = reflect.ValueOf(dest)
	c.withReturning = true

	return c
}

func (c *command) Exec(ctx context.Context) error {
	if c.src.Kind() != reflect.Pointer {
		return errors.New("model must be pointer")
	}

	c.src = c.src.Elem()

	if c.src.Kind() != reflect.Struct {
		return errors.New("model must be struct")
	}

	err := c.client.registerModel(c.src.Type())
	if err != nil {
		return err
	}

	sqlFunc, err := c.client.getSqlFunc(c.src.Type(), c.sql)
	if err != nil {
		return err
	}

	sql, sqlArgs, err := sqlFunc(c.src, c.args)
	if err != nil {
		return err
	}

	qm := c.getQueryManager(ctx)

	if !c.withReturning {
		_, err := qm.Exec(ctx, sql, sqlArgs...)
		if err != nil {
			return err
		}
	} else {
		rows, err := qm.Query(ctx, sql, sqlArgs...)
		if err != nil {
			return err
		}

		return c.client.mapRowsToDest(rows, c.dest.Elem())
	}

	return nil
}

func (c *command) getQueryManager(ctx context.Context) queryManager {
	tx, ok := getTxFromContext(ctx)
	if ok {
		return tx
	}

	return c.client.pool
}
