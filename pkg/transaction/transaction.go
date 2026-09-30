package transaction

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrNoTransaction = errors.New("no transaction in context")

type Manager interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ctxKey struct{}

func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

// ExtractDB returns the transaction from the context, or fallback when the
// call runs outside a transaction. Repositories must use it for every query.
func ExtractDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return tx
	}

	return fallback
}

func ExtractTx(ctx context.Context) (*gorm.DB, error) {
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return tx, nil
	}

	return nil, ErrNoTransaction
}
