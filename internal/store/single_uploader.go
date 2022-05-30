package store

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	fsmodels "github.com/mspraggs/hoard/internal/store/models"
	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/single_uploader.go -package=mocks -source=$GOFILE

// SingleClient defines the interface required to upload a file to a storage
// bucket in a single operation.
type SingleClient interface {
	PutObject(
		ctx context.Context,
		input *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
}

// SingleUploader encapsulates the logic to upload a file to a storage bucket in
// a single operation.
type SingleUploader struct {
	client SingleClient
	log    *zap.SugaredLogger
}

// NewSingleUploader instantiates a SingleUploader instance with the provided
// client.
func NewSingleUploader(client SingleClient) *SingleUploader {
	log := util.MustNewLogger()
	return &SingleUploader{client, log}
}

// Upload uploads the contents of the supplied file upload to the relevant
// storage bucket.
func (u *SingleUploader) Upload(
	ctx context.Context,
	upload *fsmodels.FileUpload,
) error {

	defer reportElapsedFileUploadTime(u.log, time.Now(), upload)

	u.log.Infow(
		"Uploading file with single-uploader",
		"key", upload.Key,
	)

	input := upload.ToPutObjectInput()

	_, err := u.client.PutObject(ctx, (*s3.PutObjectInput)(input))
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}

	return nil
}