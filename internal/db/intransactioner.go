package db

import (
	"context"

	"github.com/doug-martin/goqu"
)

// TxnFunc defines the signature of the closure expected by an InTransactioner
// instance. The closure accepts a Store instance, on which it can call methods
// to interact with the database.
type TxnFunc func(context.Context, *Store) (interface{}, error)

// InTransactioner is responsible for managing access to a database via a single
// transaction.
type InTransactioner struct {
	db *goqu.Database
}

// NewInTransactioner constructs a new InTransactioner instance, taking a
// database instance in order to do so.
func NewInTransactioner(db *goqu.Database) *InTransactioner {
	return &InTransactioner{db}
}

// InTransaction evaluates the provided TxnFunc within a single database
// transaction, passing it a Store instance containing the wrapped transaction.
// Errors returned by code within the provided function will result in any
// operations performed by the function being rolled back.
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
