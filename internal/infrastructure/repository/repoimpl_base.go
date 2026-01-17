package repository

import (
	"context"
	"fmt"
	"kiwi-user/internal/infrastructure/repository/ent"

	"github.com/futurxlab/golanggraph/xerror"
)

type baseImpl struct {
	db *Client
}

func (b *baseImpl) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) (err error) {

	tx, err := b.db.Tx(ctx)
	if err != nil {
		return xerror.Wrap(err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	ctx = ent.NewContext(ctx, tx.Client())

	if err := fn(ctx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return xerror.Wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return xerror.Wrap(fmt.Errorf("committing transaction: %w", err))
	}

	return nil

}

func (b *baseImpl) getEntClient(ctx context.Context) *ent.Client {
	db := ent.FromContext(ctx)

	if db == nil {
		db = b.db.Client
	}

	return db
}
