package db

import (
	"time"

	"github.com/mspraggs/hoard/internal/processor"
)

// Checksum defines a CRC32 checksum as an unsigned 32-bit integer.
type Checksum uint32

// FileRow is the database representation of a file.
type FileRow struct {
	ID                 string    `db:"id"`
	Key                string    `db:"key"`
	LocalPath          string    `db:"local_path"`
	Checksum           Checksum  `db:"checksum"`
	Bucket             string    `db:"bucket"`
	ETag               string    `db:"etag"`
	Version            string    `db:"version"`
	CreatedAtTimestamp time.Time `db:"created_at_timestamp"`
}

func (r *FileRow) toDomain() *processor.File {
	return &processor.File{
		Key:       r.Key,
		LocalPath: r.LocalPath,
		Checksum:  r.Checksum.toDomain(),
		Bucket:    r.Bucket,
		ETag:      r.ETag,
		Version:   r.Version,
	}
}

func newFileRowFromDomain(id string, file *processor.File) *FileRow {
	return &FileRow{
		ID:        id,
		Key:       file.Key,
		LocalPath: file.LocalPath,
		Checksum:  newChecksumFromDomain(file.Checksum),
		Bucket:    file.Bucket,
	}
}

func newChecksumFromDomain(checksum processor.Checksum) Checksum {
	return Checksum(checksum)
}

func (c Checksum) toDomain() processor.Checksum {
	return processor.Checksum(c)
}
