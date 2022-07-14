package db

import (
	"context"

	"github.com/mspraggs/hoard/internal/processor"
)

// Create inserts the provided file into the registry database with an ID
// generated using the ID generator provided in the registry constructor.
func (r *Registry) Create(ctx context.Context, file *processor.File) (*processor.File, error) {
	fileRow := newFileRowFromDomain(
		r.idGen.GenerateID(),
		file,
	)

	var createdFileRow *FileRow
	err := r.inTxner.InTransaction(ctx, func(ctx context.Context, tx Tx) error {
		fileRow.CreatedAtTimestamp = r.clock.Now()

		var err error
		createdFileRow, err = r.creator.Create(ctx, tx, fileRow)
		return err
	})
	if err != nil {
		return nil, err
	}

	return createdFileRow.toDomain(), nil
}
