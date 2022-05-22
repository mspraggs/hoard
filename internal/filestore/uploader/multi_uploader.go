package uploader

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
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
	maxChunkSize int64
	numChunks    int
	client       MultiClient
	log          *zap.SugaredLogger
}

// NewMultiUploader instantiates a new MultiUploader instance using the provided
// chunk size, number of chunks and client.
func NewMultiUploader(chunkSize int64, numChunks int, client MultiClient) *MultiUploader {
	log := util.MustNewLogger()
	return &MultiUploader{chunkSize, numChunks, client, log}
}

// Upload uploads the contents of the provided file upload to the relevant
// bucket in multiple parts.
func (u *MultiUploader) Upload(
	ctx context.Context,
	upload *fsmodels.FileUpload,
) error {

	defer reportElapsedFileUploadTime(u.log, time.Now(), upload)

	u.log.Infow(
		"Uploading file with multi-uploader",
		"key", upload.Key,
		"num_parts", u.numChunks,
	)

	uploadID, err := u.createMultiPartUpload(ctx, upload)
	if err != nil {
		return err
	}

	for i := 0; i < u.numChunks; i++ {
		u.log.Debugw(
			"Upload part start",
			"key", upload.Key,
			"part", i,
		)
		if err := u.uploadPart(ctx, uploadID, upload); err != nil {
			return fmt.Errorf("unable to upload part: %w", err)
		}
		u.log.Debugw(
			"Upload part finish",
			"key", upload.Key,
			"part", i,
		)
	}

	return u.closeMultiPartUpload(ctx, uploadID, upload)
}

func (u *MultiUploader) createMultiPartUpload(
	ctx context.Context,
	upload *fsmodels.FileUpload,
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
	upload *fsmodels.FileUpload,
) error {

	input := upload.ToUploadPartInput(uploadID, u.maxChunkSize)

	_, err := u.client.UploadPart(ctx, (*s3.UploadPartInput)(input))
	if err != nil {
		return fmt.Errorf("unable to upload multipart part: %w", err)
	}

	return nil
}

func (u *MultiUploader) closeMultiPartUpload(
	ctx context.Context,
	uploadID string,
	upload *fsmodels.FileUpload,
) error {

	input := upload.ToCompleteMultipartUploadInput(uploadID)

	_, err := u.client.CompleteMultipartUpload(ctx, (*s3.CompleteMultipartUploadInput)(input))
	return err
}
