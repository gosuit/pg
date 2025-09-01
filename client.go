package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Client interface {
	Query(sql string, dest any) Query
	Command(sql string, src any) Command
	Transactional(ctx context.Context, fn TxFunc, opts ...TxOption) error

	ToPgx() *pgxpool.Pool
	ToDB() *sql.DB
}

type Config struct {
	Host     string `confy:"host"     yaml:"host"     json:"host"     toml:"host"     env:"PG_HOST"    default:"localhost"`
	Port     int    `confy:"port"     yaml:"port"     json:"port"     toml:"port"     env:"PG_PORT"    default:"5432"`
	DBName   string `confy:"dbname"   yaml:"dbname"   json:"dbname"   toml:"dbname"   env:"PG_NAME"    default:"postgres"`
	Username string `confy:"username" yaml:"username" json:"username" toml:"username" env:"PG_USER"`
	Password string `confy:"password" env:"POSTGRES_PASSWORD"`
	SSLMode  string `confy:"sslmode"  yaml:"sslmode" json:"sslmode"  toml:"sslmode"   env:"PG_SSLMODE" default:"disable"`
}

func New(ctx context.Context, cfg *Config) (Client, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	))
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &client{
		pool:   pool,
		models: make(map[reflect.Type]*parsedModel),
	}, nil
}

type client struct {
	pool    *pgxpool.Pool
	models  map[reflect.Type]*parsedModel
	modelMu sync.Mutex
}

func (c *client) Query(sql string, dest any) Query {
	return &query{
		client: c,
		sql:    sql,
		dest:   reflect.ValueOf(dest),
		args:   make(map[string]any),
	}
}

func (c *client) Command(sql string, src any) Command {
	return &command{
		client:        c,
		sql:           sql,
		src:           reflect.ValueOf(src),
		withReturning: false,
		args:          make(map[string]any),
	}
}

func (c *client) Transactional(ctx context.Context, fn TxFunc, opts ...TxOption) error {
	tx, err := c.pool.BeginTx(ctx, *getTxOptions(opts))
	if err != nil {
		return err
	}

	tc := &txContext{
		tx:   tx,
		base: ctx,
	}

	if err := fn(tc); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (c *client) ToPgx() *pgxpool.Pool {
	return c.pool
}

func (c *client) ToDB() *sql.DB {
	return stdlib.OpenDBFromPool(c.pool)
}

func (c *client) registerModel(modelType reflect.Type) error {
	_, ok := c.models[modelType]
	if ok {
		return nil
	}

	c.modelMu.Lock()
	defer c.modelMu.Unlock()

	_, ok = c.models[modelType]
	if ok {
		return nil
	}

	fields, err := parseModel(modelType)
	if err != nil {
		return err
	}

	model := &parsedModel{
		fields:  fields,
		queries: make(map[string]sqlFunc),
	}

	c.models[modelType] = model

	return nil
}

func (c *client) getSqlFunc(modelType reflect.Type, sql string) sqlFunc {
	fn, ok := c.models[modelType].queries[sql]
	if ok {
		return fn
	}

	c.modelMu.Lock()
	defer c.modelMu.Unlock()

	fn, ok = c.models[modelType].queries[sql]
	if ok {
		return fn
	}

	parsedSql, keys := extractKeys(sql)

	fn = getSqlFunc(parsedSql, keys, c.models[modelType].fields)
	c.models[modelType].queries[sql] = fn

	return fn
}

func (c *client) mapRowsToDest(rows pgx.Rows, dest reflect.Value) error {
	if dest.Kind() == reflect.Struct {
		rowsCount := 0
		setters := c.models[dest.Type()].fields.setters
		for rows.Next() {
			if rowsCount > 0 {
				return errors.New("to many values")
			}

			values, err := rows.Values()
			if err != nil {
				return err
			}

			descriptions := rows.FieldDescriptions()

			for i := range descriptions {
				column := descriptions[i].Name

				setters[column](dest, reflect.ValueOf(values[i]))
			}

			rowsCount++
		}

		if rowsCount == 0 {
			return errors.New("not found value")
		}
	} else {
		modelType := dest.Type().Elem()
		if dest.Type().Elem().Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}

		setters := c.models[modelType].fields.setters
		for rows.Next() {

			values, err := rows.Values()
			if err != nil {
				return err
			}

			descriptions := rows.FieldDescriptions()

			model := reflect.New(modelType).Elem()

			for i := range descriptions {
				column := descriptions[i].Name

				setters[column](model, reflect.ValueOf(values[i]))
			}

			if dest.Type().Elem().Kind() == reflect.Pointer {
				model = model.Addr()
			}

			dest.Set(reflect.Append(dest, model))
		}
	}

	return nil
}
