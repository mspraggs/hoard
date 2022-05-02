package uploader

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
)

//go:generate mockgen -destination=./mocks/single_uploader.go -package=mocks -source=$GOFILE

type SingleClient interface {
	PutObject(
		ctx context.Context,
		input *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
}

type SingleUploader struct {
	client SingleClient
}

func NewSingleUploader(client SingleClient) *SingleUploader {
	return &SingleUploader{client}
}

func (u *SingleUploader) Upload(
	ctx context.Context,
	reader io.Reader,
	upload *fsmodels.FileUpload,
) error {

	setBody := func(i *fsmodels.PutObjectInput) {
		i.Body = reader
	}

	input := upload.ToPutObjectInput(setBody)

	_, err := u.client.PutObject(ctx, (*s3.PutObjectInput)(input))
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}

	return nil
}
