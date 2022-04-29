package models

import "time"

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

type EncryptionAlgorithm int

const (
	EncryptionAlgorithmAES256 EncryptionAlgorithm = 1
)

type EncryptionKey string

type ChecksumAlgorithm int

const (
	ChecksumAlgorithmSHA256 ChecksumAlgorithm = 1
)

type Checksum string

type ChangeType int

const (
	ChangeTypeCreate ChangeType = 1
	ChangeTypeUpdate ChangeType = 2
)
