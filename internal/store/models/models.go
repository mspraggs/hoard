package models

import (
	"crypto/md5"
	"encoding/base64"
	"io"

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
	Size                int64
	EncryptionKey       EncryptionKey
	EncryptionAlgorithm EncryptionAlgorithm
	ChecksumAlgorithm   ChecksumAlgorithm
	StorageClass        StorageClass
	Body                io.Reader
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

// UploadPartOutput defines the input data returned after uploading one part of
// a multi-part upload.
type UploadPartOutput s3.UploadPartOutput

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

// DeleteObjectInput defines the input data required to delete a file upload.
type DeleteObjectInput s3.DeleteObjectInput

// NewEncryptionKeyFromBusiness creates a store EncryptionKey from a
// business EncryptionKey.
func NewEncryptionKeyFromBusiness(key models.EncryptionKey) EncryptionKey {
	return EncryptionKey(key)
}

// NewEncryptionAlgorithmFromBusiness creates a store EncryptionAlgorithm
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

// NewChecksumAlgorithmFromBusiness creates a store ChecksumAlgorithm from a
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

// NewStorageClassFromBusiness creates a store StorageClass from a business
// StorageClass.
func NewStorageClassFromBusiness(sc models.StorageClass) StorageClass {
	switch sc {
	case models.StorageClassStandard:
		return types.StorageClassStandard
	case models.StorageClassArchiveFlexi:
		return types.StorageClassGlacier
	case models.StorageClassArchiveDeep:
		return types.StorageClassDeepArchive
	case models.StorageClassArchiveInstant:
		return types.StorageClassGlacierIr
	default:
		return StorageClass("")
	}
}

// NewFileUploadFromBusiness creates a store file upload model from the
// provided business file upload, encryption key and checksum algorithm.
func NewFileUploadFromBusiness(
	encryptionAlgorithm models.EncryptionAlgorithm,
	encryptionKey models.EncryptionKey,
	checksumAlgorithm models.ChecksumAlgorithm,
	storageClass models.StorageClass,
	size int64,
	upload *models.FileUpload,
	body io.Reader,
) *FileUpload {

	return &FileUpload{
		Key:                 upload.ID,
		Bucket:              upload.Bucket,
		Size:                size,
		EncryptionKey:       NewEncryptionKeyFromBusiness(encryptionKey),
		EncryptionAlgorithm: NewEncryptionAlgorithmFromBusiness(encryptionAlgorithm),
		ChecksumAlgorithm:   NewChecksumAlgorithmFromBusiness(checksumAlgorithm),
		StorageClass:        NewStorageClassFromBusiness(storageClass),
		Body:                body,
	}
}

// ToCreateMultipartUploadInput constructs a multi-part upload creation input
// from the file upload this method is called on.
func (fu *FileUpload) ToCreateMultipartUploadInput() *CreateMultipartUploadInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	sseKeyMD5 := md5Hash(fu.EncryptionKey)
	input := &CreateMultipartUploadInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
		StorageClass:         fu.StorageClass,
	}

	return input
}

// ToUploadPartInput constructs an UploadPartInput from the file upload this
// method is called on.
func (fu *FileUpload) ToUploadPartInput(
	uploadID string,
	chunkNum int32,
	chunkSize int64,
) *UploadPartInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	sseKeyMD5 := md5Hash(fu.EncryptionKey)
	input := &UploadPartInput{
		UploadId:             &uploadID,
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		PartNumber:           chunkNum,
		ContentLength:        chunkSize,
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
		Body:                 &io.LimitedReader{R: fu.Body, N: chunkSize},
	}

	return input
}

// ToCompleteMultipartUploadInput constructs an CompleteMultipartUploadInput
// from the file upload this method is called on.
func (fu *FileUpload) ToCompleteMultipartUploadInput(
	uploadID string,
	parts []*UploadPartOutput,
) *CompleteMultipartUploadInput {

	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = types.CompletedPart{
			ChecksumCRC32:  part.ChecksumCRC32,
			ChecksumCRC32C: part.ChecksumCRC32C,
			ChecksumSHA1:   part.ChecksumSHA1,
			ChecksumSHA256: part.ChecksumSHA256,
			ETag:           part.ETag,
			PartNumber:     int32(i + 1),
		}
	}

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	sseKeyMD5 := md5Hash(fu.EncryptionKey)
	input := &CompleteMultipartUploadInput{
		UploadId: &uploadID,
		Bucket:   &fu.Bucket,
		Key:      &fu.Key,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
	}

	return input
}

// ToPutObjectInput constructs an PutObjectInput from the file upload this
// method is called on.
func (fu *FileUpload) ToPutObjectInput() *PutObjectInput {

	sseKey := base64.StdEncoding.EncodeToString(fu.EncryptionKey)
	sseKeyMD5 := md5Hash(fu.EncryptionKey)
	input := &PutObjectInput{
		Bucket:               &fu.Bucket,
		Key:                  &fu.Key,
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: (*string)(&fu.EncryptionAlgorithm),
		ChecksumAlgorithm:    fu.ChecksumAlgorithm,
		StorageClass:         fu.StorageClass,
		Body:                 fu.Body,
	}

	return input
}

// ToDeleteObjectInput constructs a DeleteObjectInput from the file upload this
// method is called on.
func (fu *FileUpload) ToDeleteObjectInput() *DeleteObjectInput {
	return &DeleteObjectInput{
		Bucket: &fu.Bucket,
		Key:    &fu.Key,
	}
}

func md5Hash(input []byte) string {
	hash := md5.Sum(input)
	return base64.StdEncoding.EncodeToString(hash[:])
}
