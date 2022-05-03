package models

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/mspraggs/hoard/internal/models"
)

// StorageClass denotes a particular storage class, as specified by the storage
// backend.
type StorageClass = types.StorageClass

// EncryptionKey defines an encryption key as a sequence of bytes.
type EncryptionKey []byte

// EncryptionAlgorithm denotes the encryption algorithm used by the storage
// backend when encrypting files.
type EncryptionAlgorithm = types.ServerSideEncryption

// ChecksumAlgorithm defines the algorithm used to generate a checksum when
// verifying uploads to the storage backend.
type ChecksumAlgorithm = types.ChecksumAlgorithm

// FileUpload encapsulates all information necessary to upload a file to a
// storage backend.
type FileUpload struct {
	Key                 string
	Bucket              string
	EncryptionKey       EncryptionKey
	EncryptionAlgorithm EncryptionAlgorithm
	ChecksumAlgorithm   ChecksumAlgorithm
	StorageClass        StorageClass
}

// CreateMultipartUploadInput defines the input data required to initiate a
// multi-part file upload.
type CreateMultipartUploadInput s3.CreateMultipartUploadInput

// CreateMultipartUploadInputOption defines a mechanism to manipulate multi-part
// upload creation input objects.
type CreateMultipartUploadInputOption func(*CreateMultipartUploadInput)

// UploadPartInput defines the input data required to upload one part of a
// multi-part upload.
type UploadPartInput s3.UploadPartInput

// UploadPartInputOption defines a mechansim to manipulate multi-part part
// upload input objects.
type UploadPartInputOption func(*UploadPartInput)

// CompleteMultipartUploadInput defines the input data required to finalise a
// multi-part file upload.
type CompleteMultipartUploadInput s3.CompleteMultipartUploadInput

// CompleteMultipartUploadInputOption defines a mechansim to manipulate
// multi-part upload completion input objects.
type CompleteMultipartUploadInputOption func(*CompleteMultipartUploadInput)

// PutObjectInput defines the input required to upload an object to a storage
// backend.
type PutObjectInput s3.PutObjectInput

// PutObjectInputOption defines a mechanism to manipulate PutObjectInput
// instances.
type PutObjectInputOption func(*PutObjectInput)

// NewEncryptionKeyFromBusiness creates a filestore EncryptionKey from a
// business EncryptionKey.
func NewEncryptionKeyFromBusiness(key models.EncryptionKey) EncryptionKey {
	return EncryptionKey(key)
}

// NewEncryptionAlgorithmFromBusiness creates a filestore EncryptionAlgorithm
// from a business EncryptionAlgorithm.
func NewEncryptionAlgorithmFromBusiness(alg models.EncryptionAlgorithm) EncryptionAlgorithm {
	switch alg {
	case models.EncryptionAlgorithmAES256:
		return types.ServerSideEncryptionAes256
	default:
		// TODO: Revisit default behaviour.
		return types.ServerSideEncryptionAes256
	}
}

// NewChecksumAlgorithmFromBusiness creates a filestore ChecksumAlgorithm from a
// business ChecksumAlgorithm.
func NewChecksumAlgorithmFromBusiness(alg models.ChecksumAlgorithm) ChecksumAlgorithm {
	switch alg {
	case models.ChecksumAlgorithmSHA256:
		return types.ChecksumAlgorithmSha256
	default:
		// TODO: Revisit default behaviour.
		return types.ChecksumAlgorithmSha256
	}
}

// NewFileUploadFromBusiness creates a filestore file upload model from the
// provided business file upload, encryption key and checksum algorithm.
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

// ToCreateMultipartUploadInput constructs a multi-part upload creation input
// from the file upload this method is called on, applying the various options
// to the new input before returning it.
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

// ToUploadPartInput constructs an UploadPartInput from the file upload this
// method is called on, applying the various options to the new input before
// returning it.
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

// ToCompleteMultipartUploadInput constructs an CompleteMultipartUploadInput
// from the file upload this method is called on, applying the various options
// to the new input before returning it.
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

// ToPutObjectInput constructs an PutObjectInput from the file upload this
// method is called on, applying the various options to the new input before
// returning it.
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
