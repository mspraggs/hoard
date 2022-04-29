package models

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/mspraggs/hoard/internal/models"
)

type StorageClass = types.StorageClass

type EncryptionKey []byte

type EncryptionAlgorithm = types.ServerSideEncryption

type ChecksumAlgorithm = types.ChecksumAlgorithm

type FileUpload struct {
	Key                 string
	Bucket              string
	EncryptionKey       EncryptionKey
	EncryptionAlgorithm EncryptionAlgorithm
	ChecksumAlgorithm   ChecksumAlgorithm
	StorageClass        StorageClass
}

type CreateMultipartUploadInput s3.CreateMultipartUploadInput

type CreateMultipartUploadInputOption func(*CreateMultipartUploadInput)

type UploadPartInput s3.UploadPartInput

type UploadPartInputOption func(*UploadPartInput)

type CompleteMultipartUploadInput s3.CompleteMultipartUploadInput

type CompleteMultipartUploadInputOption func(*CompleteMultipartUploadInput)

type PutObjectInput s3.PutObjectInput

type PutObjectInputOption func(*PutObjectInput)

func NewEncryptionKeyFromBusiness(key models.EncryptionKey) EncryptionKey {
	return EncryptionKey(key)
}

func NewEncryptionAlgorithmFromBusiness(alg models.EncryptionAlgorithm) EncryptionAlgorithm {
	switch alg {
	case models.EncryptionAlgorithmAES256:
		return types.ServerSideEncryptionAes256
	default:
		// TODO: Revisit default behaviour.
		return types.ServerSideEncryptionAes256
	}
}

func NewChecksumAlgorithmFromBusiness(alg models.ChecksumAlgorithm) ChecksumAlgorithm {
	switch alg {
	case models.ChecksumAlgorithmSHA256:
		return types.ChecksumAlgorithmSha256
	default:
		// TODO: Revisit default behaviour.
		return types.ChecksumAlgorithmSha256
	}
}

func NewFileUploadFromBusiness(
	encryptionAlgorithm models.EncryptionAlgorithm,
	encryptionKey models.EncryptionKey,
	checksumAlgorithm models.ChecksumAlgorithm,
	upload *models.FileUpload,
) *FileUpload {

	return &FileUpload{
		Key:                 upload.ID,
		Bucket:              upload.Bucket,
		EncryptionKey:       NewEncryptionKeyFromBusiness(encryptionKey),
		EncryptionAlgorithm: NewEncryptionAlgorithmFromBusiness(encryptionAlgorithm),
		ChecksumAlgorithm:   NewChecksumAlgorithmFromBusiness(checksumAlgorithm),
	}
}

func (fu *FileUpload) ToCreateMultipartUploadInput(
	opts ...CreateMultipartUploadInputOption,
) *CreateMultipartUploadInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	input := &CreateMultipartUploadInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
		StorageClass:         fu.StorageClass,
	}

	for _, opt := range opts {
		opt(input)
	}

	return input
}

func (fu *FileUpload) ToUploadPartInput(
	opts ...UploadPartInputOption,
) *UploadPartInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	input := &UploadPartInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
	}

	for _, opt := range opts {
		opt(input)
	}

	return input
}

func (fu *FileUpload) ToCompleteMultipartUploadInput(
	opts ...CompleteMultipartUploadInputOption,
) *CompleteMultipartUploadInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	input := &CompleteMultipartUploadInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
	}

	for _, opt := range opts {
		opt(input)
	}

	return input
}

func (fu *FileUpload) ToPutObjectInput(
	opts ...PutObjectInputOption,
) *PutObjectInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	input := &PutObjectInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
		StorageClass:         fu.StorageClass,
	}

	for _, opt := range opts {
		opt(input)
	}

	return input
}

func (upi *UploadPartInput) AttachChecksum(checksum models.Checksum) error {
	switch upi.ChecksumAlgorithm {
	case types.ChecksumAlgorithmSha256:
		upi.ChecksumSHA256 = (*string)(&checksum)
	default:
		return fmt.Errorf("unknown checksum algorithm: %s", upi.ChecksumAlgorithm)
	}
	return nil
}

func (poi *PutObjectInput) AttachChecksum(checksum models.Checksum) error {
	switch poi.ChecksumAlgorithm {
	case types.ChecksumAlgorithmSha256:
		poi.ChecksumSHA256 = (*string)(&checksum)
	default:
		return fmt.Errorf("unknown checksum algorithm: %s", poi.ChecksumAlgorithm)
	}
	return nil
}
