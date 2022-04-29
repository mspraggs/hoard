package uploader

import (
	"bytes"
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
	cs Checksummer,
	upload *fsmodels.FileUpload,
) error {

	uploadID, err := u.createMultiPartUpload(ctx, upload)
	if err != nil {
		return err
	}

	for {
		n, err := u.uploadPart(ctx, reader, cs, uploadID, upload)
		if err != nil {
			return fmt.Errorf("unable to upload part: %w", err)
		}
		if n == 0 {
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
	cs Checksummer,
	uploadID string,
	upload *fsmodels.FileUpload,
) (int, error) {

	buffer := make([]byte, u.maxChunkSize)
	n, err := reader.Read(buffer)
	if n == 0 {
		return 0, err
	}
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("unable to read data: %w", err)
	}

	bufferReader := bytes.NewReader(buffer[:n])
	checksum, err := cs.Checksum(bufferReader)
	if err != nil {
		return 0, fmt.Errorf("unable to calculate checksum: %w", err)
	}

	bufferReader.Reset(buffer[:n])

	setUploadID := func(i *fsmodels.UploadPartInput) {
		i.UploadId = &uploadID
	}
	setBody := func(i *fsmodels.UploadPartInput) {
		i.Body = bufferReader
	}
	setChecksum := func(i *fsmodels.UploadPartInput) {
		i.AttachChecksum(checksum)
	}

	input := upload.ToUploadPartInput(setUploadID, setBody, setChecksum)

	_, err = u.client.UploadPart(ctx, (*s3.UploadPartInput)(input))
	if err != nil {
		return 0, fmt.Errorf("unable to upload multipart part: %w", err)
	}

	return n, nil
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
