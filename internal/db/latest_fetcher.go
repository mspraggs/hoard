package db

import (
	"context"
	"database/sql"
)

const getLatestFile = `-- name: GetLatestFile :one
SELECT
	id,
	key,
	local_path,
	checksum,
	bucket,
	etag,
	version,
	salt,
	encryption_algorithm,
	key_params,
	created_at_timestamp
FROM files
WHERE local_path = $1
ORDER BY created_at_timestamp DESC
LIMIT 1
`

// GoquLatestFetcher provides the logic to fetch the most recent version of a
// file with a given path within a transaction.
type GoquLatestFetcher struct{}

// NewGoquLatestFetcher instantiates a new GoLatestFetcher instance.
func NewGoquLatestFetcher() *GoquLatestFetcher {
	return &GoquLatestFetcher{}
}

// FetchLatest returns the most recent version of a file with the provided path.
func (lf *GoquLatestFetcher) FetchLatest(
	ctx context.Context,
	tx Tx,
	path string,
) (*FileRow, error) {

	row := tx.QueryRowContext(ctx, getLatestFile, path)

	var selectedFile FileRow
	if err := row.Scan(
		&selectedFile.ID,
		&selectedFile.Key,
		&selectedFile.LocalPath,
		&selectedFile.Checksum,
		&selectedFile.Bucket,
		&selectedFile.ETag,
		&selectedFile.Version,
		&selectedFile.Salt,
		&selectedFile.EncryptionAlgorithm,
		&selectedFile.KeyParams,
		&selectedFile.CreatedAtTimestamp,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &selectedFile, nil
}