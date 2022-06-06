package store

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/client.go -package=mocks -source=$GOFILE

// BackendClient is the interface required to interact with a storage backend.
type BackendClient interface {
	PutObject(
		ctx context.Context,
		input *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
	CreateMultipartUpload(
		ctx context.Context,
		input *s3.CreateMultipartUploadInput,
		optFns ...func(*s3.Options),
	) (*s3.CreateMultipartUploadOutput, error)
	UploadPart(
		ctx context.Context,
		input *s3.UploadPartInput,
		optFns ...func(*s3.Options),
	) (*s3.UploadPartOutput, error)
	CompleteMultipartUpload(
		ctx context.Context,
		input *s3.CompleteMultipartUploadInput,
		optFns ...func(*s3.Options),
	) (*s3.CompleteMultipartUploadOutput, error)
	DeleteObject(
		ctx context.Context,
		input *s3.DeleteObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
}

// AWSClient provides a thin wrapper around the required BackendClient interface to
// support single- and multi-part file upload.
type AWSClient struct {
	chunksize int64
	bc        BackendClient
	log       *zap.SugaredLogger
}

// NewClient constructs a new client instance using the provided backend client
// and chunk size.
func NewAWSClient(chunksize int64, bc BackendClient) *AWSClient {
	return &AWSClient{
		chunksize: chunksize,
		bc:        bc,
		log:       util.MustNewLogger(),
	}
}
