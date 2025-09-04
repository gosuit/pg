package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type TxFunc func(ctx context.Context) error
type TxOption func(opts *pgx.TxOptions)

type TxIsoLevel = pgx.TxIsoLevel
type TxAccessMode = pgx.TxAccessMode
type TxDeferrableMode = pgx.TxDeferrableMode

const (
	Serializable    TxIsoLevel = pgx.Serializable
	RepeatableRead  TxIsoLevel = pgx.RepeatableRead
	ReadCommitted   TxIsoLevel = pgx.ReadCommitted
	ReadUncommitted TxIsoLevel = pgx.ReadUncommitted

	ReadWrite TxAccessMode = pgx.ReadWrite
	ReadOnly  TxAccessMode = pgx.ReadOnly

	Deferrable    TxDeferrableMode = pgx.Deferrable
	NotDeferrable TxDeferrableMode = pgx.NotDeferrable
)

func WithLevel(level TxIsoLevel) TxOption {
	return func(opts *pgx.TxOptions) {
		opts.IsoLevel = level
	}
}

func WithAccess(access TxAccessMode) TxOption {
	return func(opts *pgx.TxOptions) {
		opts.AccessMode = access
	}
}

func WithDeferrable(deferrable TxDeferrableMode) TxOption {
	return func(opts *pgx.TxOptions) {
		opts.DeferrableMode = deferrable
	}
}

func WithBeginQuery(query string) TxOption {
	return func(opts *pgx.TxOptions) {
		opts.BeginQuery = query
	}
}

func WithCommitQuery(query string) TxOption {
	return func(opts *pgx.TxOptions) {
		opts.CommitQuery = query
	}
}

func getTxOptions(opts []TxOption) *pgx.TxOptions {
	result := &pgx.TxOptions{}

	for _, o := range opts {
		o(result)
	}

	return result
}

type txContext struct {
	tx   pgx.Tx
	base context.Context
}

func (tc *txContext) Deadline() (time.Time, bool) {
	return tc.base.Deadline()
}

func (tc *txContext) Done() <-chan struct{} {
	return tc.base.Done()
}

func (tc *txContext) Err() error {
	return tc.base.Err()
}

func (tc *txContext) Value(key any) any {
	return tc.base.Value(key)
}

func getTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	if tc, ok := ctx.(*txContext); ok {
		return tc.tx, true
	}

	return nil, false
}
