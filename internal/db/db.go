package db

import (
	"context"
	"database/sql"
)

//go:generate mockgen -destination=./mocks/db.go -package=mocks -source=$GOFILE

// Tx encapsulates a DB transaction.
type Tx interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// DB encapsulates a database
type DB interface {
	Begin() (Tx, error)
}

// TxnFunc defines the signature of the closure expected by an InTransactioner
// instance. The closure accepts a Store instance, on which it can call methods
// to interact with the database.
type TxnFunc = func(context.Context, Tx) error
