package db

import (
	"context"
)

const createFile = `-- name: CreateFile :one
INSERT INTO files.files (
	id,
	key,
	local_path,
	checksum,
	change_time,
	bucket,
	etag,
	version,
	created_at_timestamp
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, key, local_path, checksum, change_time, bucket, etag, version, created_at_timestamp
`

// CreatorTx provides the logic to insert a file into a database within a
// transaction.
type CreatorTx struct{}

// NewCreatorTx instantiates a new CreatorTx instance.
func NewCreatorTx() *CreatorTx {
	return &CreatorTx{}
}

// Create inserts the provided row into the database using the provided
// transaction.
func (c *CreatorTx) Create(
	ctx context.Context,
	tx Tx,
	file *FileRow,
) (*FileRow, error) {

	row := tx.QueryRowContext(
		ctx,
		createFile,
		file.ID,
		file.Key,
		file.LocalPath,
		file.Checksum,
		file.CTime,
		file.Bucket,
		file.ETag,
		file.Version,
		file.CreatedAtTimestamp,
	)
	var insertedFile FileRow
	err := row.Scan(
		&insertedFile.ID,
		&insertedFile.Key,
		&insertedFile.LocalPath,
		&insertedFile.Checksum,
		&insertedFile.CTime,
		&insertedFile.Bucket,
		&insertedFile.ETag,
		&insertedFile.Version,
		&insertedFile.CreatedAtTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &insertedFile, nil
}
