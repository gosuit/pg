package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host     string `confy:"host"     yaml:"host"     json:"host"     toml:"host"     env:"PG_HOST"    default:"localhost"`
	Port     int    `confy:"port"     yaml:"port"     json:"port"     toml:"port"     env:"PG_PORT"    default:"5432"`
	DBName   string `confy:"dbname"   yaml:"dbname"   json:"dbname"   toml:"dbname"   env:"PG_NAME"    default:"postgres"`
	Username string `confy:"username" yaml:"username" json:"username" toml:"username" env:"PG_USER"`
	Password string `confy:"password" env:"POSTGRES_PASSWORD"`
	SSLMode  string `confy:"sslmode"  yaml:"sslmode" json:"sslmode"  toml:"sslmode"   env:"PG_SSLMODE" default:"disable"`
}

type Client interface {
	Query(sql string, model any) Query
	Command(sql string, model any) Command
	Transactional(ctx context.Context, fn TxFunc) error

	ToPgx() *pgxpool.Pool
	ToDB() *sql.DB
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

	return &client{pool}, nil
}

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
