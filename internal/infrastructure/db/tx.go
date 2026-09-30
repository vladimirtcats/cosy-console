package db

import (
	"context"

	"gorm.io/gorm"

	"cosy-console/pkg/transaction"
)

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(transaction.InjectTx(ctx, tx))
	})
}
