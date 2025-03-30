package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

// Client is interface for communication with postgres database.
type Client interface {
	// Registers custom types with the database connection.
	RegisterTypes(types []string) error

	// Returns the underlying pgxpool.Pool instance.
	ToPgx() *pgxpool.Pool

	// Embeds the Pool interface for pool functionalities.
	Pool
}

// Config is type for database connection.
type Config struct {
	Host           string `confy:"host"     yaml:"host"     json:"host"     toml:"host"     env:"PG_HOST" env-default:"localhost"`
	Port           int    `confy:"port"     yaml:"port"     json:"port"     toml:"port"     env:"PG_PORT" env-default:"5432"`
	DBName         string `confy:"dbname"   yaml:"dbname"   json:"dbname"   toml:"dbname"   env:"PG_NAME" env-default:"postgres"`
	Username       string `confy:"username" yaml:"username" json:"username" toml:"username" env:"PG_USER"`
	Password       string `confy:"password" env:"POSTGRES_PASSWORD"`
	SSLMode        string `confy:"sslmode"         yaml:"sslmode"         json:"sslmode"         toml:"sslmode"         env:"PG_SSLMODE"         env-default:"disable"`
	MigrationsRun  bool   `confy:"migrations_run"  yaml:"migrations_run"  json:"migrations_run"  toml:"migrations_run"  env:"PG_MIGRATIONS_RUN"  env-default:"false"`
	MigrationsPath string `confy:"migrations_path" yaml:"migrations_path" json:"migrations_path" toml:"migrations_path" env:"PG_MIGRATIONS_PATH" env-default:"./migrations"`
}

type pgclient struct {
	afterConnectFuncs []func(ctx context.Context, conn *Conn) error
	*pgxpool.Pool
}

// New creates a new Client instance and establishes a connection to the database.
// It parses the provided configuration and runs migrations if specified.
func New(ctx context.Context, cfg *Config) (Client, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	))
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	if cfg.MigrationsRun {
		sql := stdlib.OpenDBFromPool(db)

		if err := sql.Ping(); err != nil {
			return nil, err
		}

		if err := goose.SetDialect("postgres"); err != nil {
			return nil, err
		}

		if err := goose.Up(sql, cfg.MigrationsPath); err != nil {
			return nil, err
		}

		if err := sql.Close(); err != nil {
			return nil, err
		}
	}

	client := &pgclient{
		Pool:              db,
		afterConnectFuncs: make([]func(ctx context.Context, conn *pgx.Conn) error, 0),
	}

	return client, nil
}

// NewWithPool creates a new Client instance using an existing pgxpool.Pool.
func NewWithPool(ctx context.Context, pool *pgxpool.Pool) (Client, error) {
	client := &pgclient{
		Pool:              pool,
		afterConnectFuncs: make([]func(ctx context.Context, conn *Conn) error, 0),
	}

	return client, nil
}

func (pc *pgclient) RegisterTypes(types []string) error {
	function := func(ctx context.Context, conn *pgx.Conn) error {
		for _, typeName := range types {
			t, err := conn.LoadType(ctx, typeName)
			if err != nil {
				return err
			}

			conn.TypeMap().RegisterType(t)
		}

		return nil
	}

	pc.afterConnectFuncs = append(pc.afterConnectFuncs, function)

	cfg := pc.Pool.Config()

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, f := range pc.afterConnectFuncs {
			if err := f(ctx, conn); err != nil {
				return err
			}
		}

		return nil
	}

	ctx := context.Background()

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return err
	}

	if err := db.Ping(ctx); err != nil {
		return err
	}

	pc.Pool = db

	return nil
}

func (pc *pgclient) ToPgx() *pgxpool.Pool {
	return pc.Pool
}
