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
	DeletedAtTimestamp  time.Time
}

// IsUploaded returns true if the file upload has been uploaded.
func (fu *FileUpload) IsUploaded() bool {
	return !fu.UploadedAtTimestamp.IsZero()
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
		return 256 / 8, nil
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

// StorageClass describes a file upload's storage class as an integer.
type StorageClass int

const (
	// StorageClassStandard denotes the standard storage backend object class.
	StorageClassStandard StorageClass = 1
	// StorageClassArchiveFlexi denotes long-term backend storage with read
	// times ranging from minutes to 12 hours.
	StorageClassArchiveFlexi StorageClass = 2
	// StorageClassArchiveDeep denotes long-term backend storage with long read
	// times between 12 and 48 hours.
	StorageClassArchiveDeep StorageClass = 3
	// StorageClassArchiveInstant denotes long-term backend storage with instant
	// access reads.
	StorageClassArchiveInstant StorageClass = 4
)
