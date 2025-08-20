package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type TxFunc func(ctx context.Context) error

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
