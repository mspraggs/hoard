package store

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mspraggs/hoard/internal/processor"
)

// StorageClass denotes a particular storage class, as specified by the storage
// backend.
type StorageClass = types.StorageClass

// ChecksumAlgorithm defines the algorithm used to generate a checksum when
// verifying uploads to the storage backend.
type ChecksumAlgorithm = types.ChecksumAlgorithm

// File encapsulates all information about a particular file required by the
// storage client.
type File struct {
	Key               string
	Bucket            string
	ChecksumAlgorithm ChecksumAlgorithm
	StorageClass      StorageClass
	File              fs.File
}

// CreateMultipartUploadInput defines the input data required to initiate a
// multi-part file.
type CreateMultipartUploadInput s3.CreateMultipartUploadInput

// UploadPartInput defines the input data required to upload one part of a
// multi-part upload.
type UploadPartInput s3.UploadPartInput

// UploadPartOutput defines the input data returned after uploading one part of
// a multi-part upload.
type UploadPartOutput s3.UploadPartOutput

// CompleteMultipartUploadInput defines the input data required to finalise a
// multi-part file.
type CompleteMultipartUploadInput s3.CompleteMultipartUploadInput

// PutObjectInput defines the input required to upload an object to a storage
// backend.
type PutObjectInput s3.PutObjectInput

// DeleteObjectInput defines the input data required to delete a file.
type DeleteObjectInput s3.DeleteObjectInput

// NewFileFromDomain creates a store file model from the provided domain file
// and checksum algorithm.
func NewFileFromDomain(
	domainFile *processor.File,
	checksumAlgorithm ChecksumAlgorithm,
	storageClass StorageClass,
	file fs.File,
) *File {

	return &File{
		Key:               domainFile.Key,
		Bucket:            domainFile.Bucket,
		ChecksumAlgorithm: checksumAlgorithm,
		StorageClass:      storageClass,
		File:              file,
	}
}

// ToCreateMultipartUploadInput constructs a multi-part upload creation input
// from the file this method is called on.
func (f *File) ToCreateMultipartUploadInput() *CreateMultipartUploadInput {

	input := &CreateMultipartUploadInput{
		Bucket:            &f.Bucket,
		Key:               &f.Key,
		ChecksumAlgorithm: f.ChecksumAlgorithm,
		StorageClass:      f.StorageClass,
	}

	return input
}

// ToUploadPartInput constructs an UploadPartInput from the file this method is
// called on.
func (f *File) ToUploadPartInput(
	uploadID string,
	chunkNum int32,
	chunkSize int64,
) *UploadPartInput {

	input := &UploadPartInput{
		UploadId:          &uploadID,
		Bucket:            &f.Bucket,
		Key:               &f.Key,
		PartNumber:        chunkNum,
		ContentLength:     chunkSize,
		ChecksumAlgorithm: f.ChecksumAlgorithm,
		Body:              &io.LimitedReader{R: f.File, N: chunkSize},
	}

	return input
}

// ToCompleteMultipartUploadInput constructs an CompleteMultipartUploadInput
// from the file this method is called on.
func (f *File) ToCompleteMultipartUploadInput(
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

	input := &CompleteMultipartUploadInput{
		UploadId: &uploadID,
		Bucket:   &f.Bucket,
		Key:      &f.Key,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	return input
}

// ToPutObjectInput constructs an PutObjectInput from the file this method is
// called on.
func (f *File) ToPutObjectInput() *PutObjectInput {

	input := &PutObjectInput{
		Bucket:            &f.Bucket,
		Key:               &f.Key,
		ChecksumAlgorithm: f.ChecksumAlgorithm,
		StorageClass:      f.StorageClass,
		Body:              f.File,
	}

	return input
}

// Size returns the size of the file.
func (f *File) Size() (int64, error) {
	info, err := f.File.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func md5Hash(input []byte) string {
	hash := md5.Sum(input)
	return base64.StdEncoding.EncodeToString(hash[:])
}
