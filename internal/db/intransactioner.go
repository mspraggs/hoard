package db

import (
	"context"

	"github.com/doug-martin/goqu"
)

type TxnFunc func(context.Context, *Store) (interface{}, error)

type InTransactioner struct {
	db *goqu.Database
}

func NewInTransactioner(db *goqu.Database) *InTransactioner {
	return &InTransactioner{db}
}

func (t *InTransactioner) InTransaction(ctx context.Context, fn TxnFunc) (interface{}, error) {
	var result interface{}
	tx, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	err = tx.Wrap(func() error {
		store := NewStore(tx)
		result, err = fn(ctx, store)
		return err
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
