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

//go:generate mockgen -destination=./mocks/multi_uploader.go -package=mocks -source=$GOFILE

// MultiClient defines the interface required to upload a file in multiple
// parts.
type MultiClient interface {
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

// MultiUploader encapsulates the logic to upload a particular file upload to a
// storage bucket in multiple parts.
type MultiUploader struct {
	fileSize     int64
	maxChunkSize int64
	client       MultiClient
	log          *zap.SugaredLogger
}

// NewMultiUploader instantiates a new MultiUploader instance using the provided
// chunk size, number of chunks and client.
func NewMultiUploader(fileSize, chunkSize int64, client MultiClient) *MultiUploader {
	log := util.MustNewLogger()
	return &MultiUploader{fileSize, chunkSize, client, log}
}

// Upload uploads the contents of the provided file upload to the relevant
// bucket in multiple parts.
func (u *MultiUploader) Upload(
	ctx context.Context,
	upload *models.FileUpload,
) error {

	defer reportElapsedFileUploadTime(u.log, time.Now(), upload)

	if u.maxChunkSize == 0 {
		return errors.New("invalid chunk size")
	}

	numChunks := int(u.fileSize / u.maxChunkSize)
	if u.fileSize%u.maxChunkSize != 0 {
		numChunks += 1
	}

	u.log.Infow(
		"Uploading file with multi-uploader",
		"key", upload.Key,
		"num_parts", numChunks,
	)

	uploadID, err := u.createMultiPartUpload(ctx, upload)
	if err != nil {
		return err
	}

	uploadOutputs := make([]*models.UploadPartOutput, numChunks)

	for i := 0; i < numChunks; i++ {
		partNum := int32(i + 1)
		u.log.Debugw(
			"Upload part start",
			"key", upload.Key,
			"part", partNum,
		)
		uploadOutput, err := u.uploadPart(ctx, uploadID, partNum, upload)
		if err != nil {
			return fmt.Errorf("unable to upload part: %w", err)
		}
		u.log.Debugw(
			"Upload part finish",
			"key", upload.Key,
			"part", partNum,
		)
		uploadOutputs[i] = uploadOutput
	}

	return u.closeMultiPartUpload(ctx, uploadID, uploadOutputs, upload)
}

func (u *MultiUploader) createMultiPartUpload(
	ctx context.Context,
	upload *models.FileUpload,
) (string, error) {

	input := upload.ToCreateMultipartUploadInput()

	output, err := u.client.CreateMultipartUpload(ctx, (*s3.CreateMultipartUploadInput)(input))
	if err != nil {
		return "", fmt.Errorf("unable to create multipart upload: %w", err)
	}

	return *output.UploadId, nil
}

func (u *MultiUploader) uploadPart(
	ctx context.Context,
	uploadID string,
	partNum int32,
	upload *models.FileUpload,
) (*models.UploadPartOutput, error) {

	chunkSize := u.maxChunkSize
	remainingBytes := u.fileSize - int64(partNum-1)*(u.maxChunkSize)
	if remainingBytes < chunkSize {
		chunkSize = remainingBytes
	}

	input := upload.ToUploadPartInput(uploadID, partNum, chunkSize)

	output, err := u.client.UploadPart(ctx, (*s3.UploadPartInput)(input))
	if err != nil {
		return nil, fmt.Errorf("unable to upload multipart part: %w", err)
	}

	return (*models.UploadPartOutput)(output), nil
}

func (u *MultiUploader) closeMultiPartUpload(
	ctx context.Context,
	uploadID string,
	parts []*models.UploadPartOutput,
	upload *models.FileUpload,
) error {

	input := upload.ToCompleteMultipartUploadInput(uploadID, parts)

	_, err := u.client.CompleteMultipartUpload(ctx, (*s3.CompleteMultipartUploadInput)(input))
	return err
}
