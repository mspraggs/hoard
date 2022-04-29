package uploader

import (
	"bytes"
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
	cs Checksummer,
	upload *fsmodels.FileUpload,
) error {

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}

	bufferReader := bytes.NewReader(buffer)
	checksum, err := cs.Checksum(bufferReader)
	if err != nil {
		return fmt.Errorf("unable to cacluate checksum: %w", err)
	}

	bufferReader.Reset(buffer)

	setBody := func(i *fsmodels.PutObjectInput) {
		i.Body = bufferReader
	}
	setChecksum := func(i *fsmodels.PutObjectInput) {
		i.AttachChecksum(checksum)
	}

	input := upload.ToPutObjectInput(setBody, setChecksum)

	_, err = u.client.PutObject(ctx, (*s3.PutObjectInput)(input))
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}

	return nil
}
