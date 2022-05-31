package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/store/models"
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

// Upload stores the contents of the provided file upload in the storage
// backend. The upload is performed either as a single put operation or a
// multi-part file upload, depending on the size of the file.
func (c *AWSClient) Upload(ctx context.Context, upload *models.FileUpload) error {
	if upload.Size < c.chunksize {
		return c.singleUpload(ctx, upload)
	}
	return c.multipartUpload(ctx, upload)
}

func (c *AWSClient) singleUpload(ctx context.Context, upload *models.FileUpload) error {

	defer c.reportElapsedFileUploadTime(time.Now(), upload)

	c.log.Infow(
		"Uploading file using single upload",
		"key", upload.Key,
	)

	input := upload.ToPutObjectInput()

	_, err := c.bc.PutObject(ctx, (*s3.PutObjectInput)(input))
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}

	return nil
}

func (c *AWSClient) multipartUpload(ctx context.Context, upload *models.FileUpload) error {

	defer c.reportElapsedFileUploadTime(time.Now(), upload)

	if c.chunksize == 0 {
		return errors.New("invalid chunk size")
	}

	numChunks := int(upload.Size / c.chunksize)
	if upload.Size%c.chunksize != 0 {
		numChunks += 1
	}

	c.log.Infow(
		"Uploading file using multi-part upload",
		"key", upload.Key,
		"num_parts", numChunks,
	)

	uploadID, err := c.createMultiPartUpload(ctx, upload)
	if err != nil {
		return err
	}

	uploadOutputs := make([]*models.UploadPartOutput, numChunks)

	for i := 0; i < numChunks; i++ {
		partNum := int32(i + 1)
		c.log.Debugw(
			"Upload part start",
			"key", upload.Key,
			"part", partNum,
		)
		uploadOutput, err := c.uploadPart(ctx, uploadID, partNum, upload)
		if err != nil {
			return fmt.Errorf("unable to upload part: %w", err)
		}
		c.log.Debugw(
			"Upload part finish",
			"key", upload.Key,
			"part", partNum,
		)
		uploadOutputs[i] = uploadOutput
	}

	return c.closeMultiPartUpload(ctx, uploadID, uploadOutputs, upload)
}

func (c *AWSClient) createMultiPartUpload(
	ctx context.Context,
	upload *models.FileUpload,
) (string, error) {

	input := upload.ToCreateMultipartUploadInput()

	output, err := c.bc.CreateMultipartUpload(ctx, (*s3.CreateMultipartUploadInput)(input))
	if err != nil {
		return "", fmt.Errorf("unable to create multipart upload: %w", err)
	}

	return *output.UploadId, nil
}

func (c *AWSClient) uploadPart(
	ctx context.Context,
	uploadID string,
	partNum int32,
	upload *models.FileUpload,
) (*models.UploadPartOutput, error) {

	chunksize := c.chunksize
	remainingBytes := upload.Size - int64(partNum-1)*(c.chunksize)
	if remainingBytes < chunksize {
		chunksize = remainingBytes
	}

	input := upload.ToUploadPartInput(uploadID, partNum, chunksize)

	output, err := c.bc.UploadPart(ctx, (*s3.UploadPartInput)(input))
	if err != nil {
		return nil, fmt.Errorf("unable to upload multipart part: %w", err)
	}

	return (*models.UploadPartOutput)(output), nil
}

func (c *AWSClient) closeMultiPartUpload(
	ctx context.Context,
	uploadID string,
	parts []*models.UploadPartOutput,
	upload *models.FileUpload,
) error {

	input := upload.ToCompleteMultipartUploadInput(uploadID, parts)

	_, err := c.bc.CompleteMultipartUpload(ctx, (*s3.CompleteMultipartUploadInput)(input))
	return err
}

func (c *AWSClient) reportElapsedFileUploadTime(
	start time.Time,
	fileUpload *models.FileUpload,
) {

	elapsed := time.Since(start)
	c.log.Infow(
		"Upload to store finished",
		"key", fileUpload.Key,
		"bucket", fileUpload.Bucket,
		"elapsed_time", elapsed,
	)
}
