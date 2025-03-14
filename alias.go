package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	// Type aliases for commonly used pgx types.

	Row            = pgx.Row
	Rows           = pgx.Rows
	Conn           = pgx.Conn
	Tx             = pgx.Tx
	TxOptions      = pgx.TxOptions
	Stat           = pgxpool.Stat
	Batch          = pgx.Batch
	BatchResults   = pgx.BatchResults
	Identifier     = pgx.Identifier
	CopyFromSource = pgx.CopyFromSource
)

var (
	// Common errors returned by the pgx package.

	ErrTxCommitRollback = pgx.ErrTxCommitRollback
	ErrTooManyRows      = pgx.ErrTooManyRows
	ErrTxClosed         = pgx.ErrTxClosed
	ErrNoRows           = pgx.ErrNoRows
)

// Pool interface defines methods for managing a pool of database connections.
type Pool interface {
	// Acquire returns a connection (*pgxpool.Conn) from the Pool
	Acquire(ctx context.Context) (c *pgxpool.Conn, err error)

	// AcquireFunc acquires a *pgxpool.Conn and calls f with that *pgxpool.Conn. ctx will only affect the Acquire. It has no effect on the
	// call of f. The return value is either an error acquiring the *pgxpool.Conn or the return value of f. The *pgxpool.Conn is
	// automatically released after the call of f.
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error

	// Exec acquires a connection from the Pool and executes the given SQL.
	// SQL can be either a prepared statement name or an SQL string.
	// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
	// The acquired connection is returned to the pool when the Exec function returns.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)

	// AcquireAllIdle atomically acquires all currently idle connections. Its intended use is for health check and
	// keep-alive functionality. It does not update pool statistics.
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn

	// QueryRow acquires a connection and executes a query that is expected
	// to return at most one row (pgx.Row). Errors are deferred until pgx.Row's
	// Scan method is called. If the query selects no rows, pgx.Row's Scan will
	// return ErrNoRows. Otherwise, pgx.Row's Scan scans the first selected row
	// and discards the rest. The acquired connection is returned to the Pool when
	// pgx.Row's Scan method is called.
	//
	// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// SendBatch sends a batch of queries to the database and returns the results.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	// Query acquires a connection and executes a query that returns pgx.Rows.
	// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
	// See pgx.Rows documentation to close the returned Rows and return the acquired connection to the Pool.
	//
	// If there is an error, the returned pgx.Rows will be returned in an error state.
	// If preferred, ignore the error returned from Query and handle errors using the returned pgx.Rows.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// BeginTx acquires a connection from the Pool and starts a transaction with pgx.TxOptions determining the transaction mode.
	// Unlike database/sql, the context only affects the begin command. i.e. there is no auto-rollback on context cancellation.
	// *pgxpool.Tx is returned, which implements the pgx.Tx interface.
	// Commit or Rollback must be called on the returned transaction to finalize the transaction block.
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)

	// CopyFrom uses the PostgreSQL copy protocol to perform bulk data insertion. It returns the number of rows copied and an error.
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)

	// Ping acquires a connection from the Pool and executes an empty sql statement against it.
	// If the sql returns without error, the database Ping is considered successful, otherwise, the error is returned.
	Ping(ctx context.Context) error

	// Begin acquires a connection from the Pool and starts a transaction. Unlike database/sql, the context only affects the begin command. i.e. there is no
	// auto-rollback on context cancellation. Begin initiates a transaction block without explicitly setting a transaction mode for the block (see BeginTx with TxOptions if transaction mode is required).
	// *pgxpool.Tx is returned, which implements the pgx.Tx interface.
	// Commit or Rollback must be called on the returned transaction to finalize the transaction block.
	Begin(ctx context.Context) (pgx.Tx, error)

	// Stat returns a pgxpool.Stat struct with a snapshot of Pool statistics.
	Stat() *pgxpool.Stat

	// Config returns a copy of config that was used to initialize this pool.
	Config() *pgxpool.Config

	// Reset closes all connections, but leaves the pool open. It is intended for use when an error is detected that would
	// disrupt all connections (such as a network interruption or a server state change).
	//
	// It is safe to reset a pool while connections are checked out. Those connections will be closed when they are returned
	// to the pool.
	Reset()

	// Close closes all connections in the pool and rejects future Acquire calls. Blocks until all connections are returned
	// to pool and closed.
	Close()
}
