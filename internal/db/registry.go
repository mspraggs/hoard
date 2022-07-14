package db

import (
	"context"
	"time"
)

//go:generate mockgen -destination=./mocks/registry.go -package=mocks -source=$GOFILE

// Clock defines the interface required to fetch the current time.
type Clock interface {
	Now() time.Time
}

// InTransactioner defines the interface required to interact with a database
// within a transaction.
type InTransactioner interface {
	InTransaction(ctx context.Context, fn TxnFunc) error
}

// LatestFetcher defines the interface required to fetch the latest version of a
// file within a database transaction.
type LatestFetcher interface {
	FetchLatest(ctx context.Context, tx Tx, path string) (*FileRow, error)
}

// Creator defines the interface required to create a file within a database
// transaction.
type Creator interface {
	Create(ctx context.Context, tx Tx, file *FileRow) (*FileRow, error)
}

// IDGenerator defines the interface required to generate a request ID for a
// given file upload.
type IDGenerator interface {
	GenerateID() string
}

// Registry encapsulates the logic required to interact with a register of
// file uploads. The registry maintains a record of details associated with a
// file upload, including a file version string and the timestamp at which the
// file was uploaded.
type Registry struct {
	clock         Clock
	idGen         IDGenerator
	inTxner       InTransactioner
	latestFetcher LatestFetcher
	creator       Creator
}

// NewRegistry instantiates a new Registry using the provided Clock,
// InTransactioner and IDGenerator instances.
func NewRegistry(
	clock Clock,
	inTxner InTransactioner,
	creator Creator,
	latestFetcher LatestFetcher,
	idGen IDGenerator,
) *Registry {

	return &Registry{
		clock:         clock,
		idGen:         idGen,
		inTxner:       inTxner,
		latestFetcher: latestFetcher,
		creator:       creator,
	}
}
