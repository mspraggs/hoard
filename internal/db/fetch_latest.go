package db

import (
	"context"

	"github.com/mspraggs/hoard/internal/processor"
)

// FetchLatest retrieves the latest version of a file with the provided path
// from the database.
func (r *Registry) FetchLatest(ctx context.Context, path string) (*processor.File, error) {
	var latestFileRow *FileRow
	err := r.inTxner.InTransaction(ctx, func(ctx context.Context, tx Tx) error {
		var err error
		latestFileRow, err = r.latestFetcher.FetchLatest(ctx, tx, path)
		return err
	})
	if err != nil {
		return nil, err
	}
	if latestFileRow == nil {
		return nil, nil
	}

	return latestFileRow.toDomain(), nil
}
