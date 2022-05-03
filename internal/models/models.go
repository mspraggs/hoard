package models

import (
	"errors"
	"time"
)

// FileUpload encapsulates all information associated with a file to be
// uploaded.
type FileUpload struct {
	ID                  string
	LocalPath           string
	Bucket              string
	Version             string
	Salt                []byte
	EncryptionAlgorithm EncryptionAlgorithm
	CreatedAtTimestamp  time.Time
	UploadedAtTimestamp time.Time
}

// EncryptionAlgorithm denotes a particular encryption algorithm using an
// integer.
type EncryptionAlgorithm int

const (
	// EncryptionAlgorithmAES256 denotes the AES256 encryption algorithm.
	EncryptionAlgorithmAES256 EncryptionAlgorithm = 1
)

// KeySize returns the key size required by a particular encryption algorithm.
func (a EncryptionAlgorithm) KeySize() (uint32, error) {
	switch a {
	case EncryptionAlgorithmAES256:
		return 256, nil
	default:
		return 0, errors.New("unknown encryption algorithm")
	}
}

// EncryptionKey defines an encryption key as a sequence of bytes.
type EncryptionKey []byte

// ChecksumAlgorithm denotes a particular checksum algorithm using an integer.
type ChecksumAlgorithm int

const (
	// ChecksumAlgorithmSHA256 denotes the SHA256 checksum algorithm.
	ChecksumAlgorithmSHA256 ChecksumAlgorithm = 1
)

// ChangeType describes a resource change type as an integer.
type ChangeType int

const (
	// ChangeTypeCreate denotes a creation change type.
	ChangeTypeCreate ChangeType = 1
	// ChangeTypeUpdate denotes a mutation change type.
	ChangeTypeUpdate ChangeType = 2
)
