package contract

import "context"

type ITransaction interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
