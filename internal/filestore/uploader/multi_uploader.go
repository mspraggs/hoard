package uploader

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
)

//go:generate mockgen -destination=./mocks/multi_uploader.go -package=mocks -source=$GOFILE

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

type MultiUploader struct {
	maxChunkSize int64
	client       MultiClient
}

func NewMultiUploader(chunkSize int64, client MultiClient) *MultiUploader {
	return &MultiUploader{chunkSize, client}
}

func (u *MultiUploader) Upload(
	ctx context.Context,
	reader io.Reader,
	upload *fsmodels.FileUpload,
) error {

	uploadID, err := u.createMultiPartUpload(ctx, upload)
	if err != nil {
		return err
	}

	for {
		limitedReader := &io.LimitedReader{R: reader, N: u.maxChunkSize}
		if err := u.uploadPart(ctx, limitedReader, uploadID, upload); err != nil {
			return fmt.Errorf("unable to upload part: %w", err)
		}
		if limitedReader.N > 0 {
			break
		}
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
	reader io.Reader,
	uploadID string,
	upload *fsmodels.FileUpload,
) error {

	setUploadID := func(i *fsmodels.UploadPartInput) {
		i.UploadId = &uploadID
	}
	setBody := func(i *fsmodels.UploadPartInput) {
		i.Body = reader
	}

	input := upload.ToUploadPartInput(setUploadID, setBody)

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

	setUploadID := func(i *fsmodels.CompleteMultipartUploadInput) {
		i.UploadId = &uploadID
	}
	input := upload.ToCompleteMultipartUploadInput(setUploadID)

	_, err := u.client.CompleteMultipartUpload(ctx, (*s3.CompleteMultipartUploadInput)(input))
	return err
}
