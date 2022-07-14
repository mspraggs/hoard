package db

import (
	"time"

	"github.com/mspraggs/hoard/internal/processor"
)

// Checksum defines a CRC32 checksum as an unsigned 32-bit integer.
type Checksum uint32

// EncryptionAlgorithm denotes a particular encryption algorithm using an
// integer.
type EncryptionAlgorithm int

const (
	// EncryptionAlgorithmAES256 denotes the AES256 encryption algorithm.
	EncryptionAlgorithmAES256 EncryptionAlgorithm = 1
)

// FileRow is the database representation of a file.
type FileRow struct {
	ID                  string              `db:"id"`
	Key                 string              `db:"key"`
	LocalPath           string              `db:"local_path"`
	Checksum            Checksum            `db:"checksum"`
	Bucket              string              `db:"bucket"`
	ETag                string              `db:"etag"`
	Version             string              `db:"version"`
	Salt                []byte              `db:"salt"`
	EncryptionAlgorithm EncryptionAlgorithm `db:"encryption_algorithm"`
	KeyParams           string              `db:"key_params"`
	CreatedAtTimestamp  time.Time           `db:"created_at_timestamp"`
}

func (r *FileRow) toDomain() *processor.File {
	// TODO: Decrypt local path
	return &processor.File{
		Key:                 r.Key,
		LocalPath:           r.LocalPath,
		Checksum:            r.Checksum.toDomain(),
		Bucket:              r.Bucket,
		ETag:                r.ETag,
		Version:             r.Version,
		Salt:                r.Salt,
		KeyParams:           r.KeyParams,
		EncryptionAlgorithm: r.EncryptionAlgorithm.toDomain(),
	}
}

func newFileRowFromDomain(id string, file *processor.File) *FileRow {
	// TODO: Encrypt local path
	return &FileRow{
		ID:                  id,
		Key:                 file.Key,
		LocalPath:           file.LocalPath,
		Checksum:            newChecksumFromDomain(file.Checksum),
		Bucket:              file.Bucket,
		Salt:                file.Salt,
		EncryptionAlgorithm: newEncryptionAlgorithmFromDomain(file.EncryptionAlgorithm),
	}
}

func newChecksumFromDomain(checksum processor.Checksum) Checksum {
	return Checksum(checksum)
}

func (c Checksum) toDomain() processor.Checksum {
	return processor.Checksum(c)
}

func newEncryptionAlgorithmFromDomain(a processor.EncryptionAlgorithm) EncryptionAlgorithm {
	switch a {
	case processor.EncryptionAlgorithmAES256:
		return EncryptionAlgorithmAES256
	default:
		return EncryptionAlgorithm(0)
	}
}

func (a EncryptionAlgorithm) toDomain() processor.EncryptionAlgorithm {
	switch a {
	case EncryptionAlgorithmAES256:
		return processor.EncryptionAlgorithmAES256
	default:
		return processor.EncryptionAlgorithm(0)
	}
}
