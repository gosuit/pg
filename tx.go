package pg

import "context"

type TxFunc func(ctx context.Context) error
