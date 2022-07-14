package db

import (
	"context"
	"database/sql"
)

//go:generate mockgen -destination=./mocks/intransactioner.go -package=mocks -source=$GOFILE

// StdInTransactioner is responsible for managing access to a database via a
// single transaction.
type StdInTransactioner struct {
	db *sql.DB
}

// NewInTransactioner constructs a new StdInTransactioner instance, taking a
// database instance in order to do so.
func NewInTransactioner(db *sql.DB) *StdInTransactioner {
	return &StdInTransactioner{db}
}

// InTransaction evaluates the provided TxnFunc within a single database
// transaction. Errors returned by code within the provided function will result
// in any operations performed by the function being rolled back.
func (t *StdInTransactioner) InTransaction(ctx context.Context, fn TxnFunc) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
