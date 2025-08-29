package pg

import (
	"context"
	"errors"
	"reflect"
	"slices"
)

type Query interface {
	WithArgs(args ...*Argument) Query
	WithArg(key string, value any) Query
	Exec(ctx context.Context) error
}

type query struct {
	client *client
	sql    string
	dest   reflect.Value
	args   map[string]any
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

var validQueryDestKinds = []reflect.Kind{reflect.Struct, reflect.Array, reflect.Slice}

func (q *query) Exec(ctx context.Context) error {
	if q.dest.Kind() != reflect.Pointer {
		return errors.New("dest must be pointer")
	}

	q.dest = q.dest.Elem()

	if !slices.Contains(validQueryDestKinds, q.dest.Kind()) || slices.Contains(specificTypes, q.dest.Type()) {
		return errors.New("dest must be struct or array")
	}

	if q.dest.Kind() != reflect.Struct {
		elemType := q.dest.Type().Elem()

		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("dest must be struct or array")
		}
	}

	var destType reflect.Type

	if q.dest.Kind() == reflect.Struct {
		destType = q.dest.Type()
	} else {
		if q.dest.Type().Elem().Kind() == reflect.Pointer {
			destType = q.dest.Type().Elem().Elem()
		} else {
			destType = q.dest.Type().Elem()
		}
	}

	err := q.client.registerModel(destType)
	if err != nil {
		return err
	}

	sqlFunc := q.client.getSqlFunc(destType, q.sql)

	sql, sqlArgs := sqlFunc(q.dest, q.args)

	qm := q.getQueryManager(ctx)

	rows, err := qm.Query(ctx, sql, sqlArgs...)
	if err != nil {
		return err
	}

	return q.client.mapRowsToDest(rows, q.dest)
}

func (q *query) getQueryManager(ctx context.Context) queryManager {
	tx, ok := getTxFromContext(ctx)
	if ok {
		return tx
	}

	return q.client.pool
}
