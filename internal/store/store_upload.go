package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/mspraggs/hoard/internal/processor"
)

// Upload stores the contents of the provided file file in the storage backend.
// The file is performed either as a single put operation or a multi-part file
// file, depending on the size of the file.
func (s *Store) Upload(ctx context.Context, file *processor.File) (*processor.File, error) {
	if s.chunksize == 0 {
		return nil, errors.New("invalid chunk size")
	}

	f, err := s.fs.Open(file.LocalPath)
	if err != nil {
		return nil, err
	}

	file.Bucket = s.bucket
	storeFile := NewFileFromDomain(file, s.csAlg, s.sc, f)

	eTag, version, err := s.upload(ctx, storeFile)
	if err != nil {
		return nil, err
	}

	file.ETag = eTag
	file.Version = version

	return file, nil
}

func (s *Store) upload(ctx context.Context, file *File) (string, string, error) {
	size, err := file.Size()
	if err != nil {
		return "", "", err
	}

	if size < s.chunksize {
		return s.singleUpload(ctx, file)
	} else {
		return s.multipartUpload(ctx, size, file)
	}
}

func (s *Store) singleUpload(ctx context.Context, file *File) (string, string, error) {
	defer s.reportElapsedFileUploadTime(time.Now(), file)

	s.log.Infow(
		"Uploading file using single file",
		"key", file.Key,
	)

	input := file.ToPutObjectInput()

	output, err := s.client.PutObject(ctx, (*s3.PutObjectInput)(input))
	if err != nil {
		return "", "", fmt.Errorf("unable to put object: %w", err)
	}

	version := ""
	if output.VersionId != nil {
		version = *output.VersionId
	}

	return *output.ETag, version, nil
}

func (s *Store) multipartUpload(
	ctx context.Context,
	size int64,
	file *File,
) (string, string, error) {

	defer s.reportElapsedFileUploadTime(time.Now(), file)

	numChunks := int(size / s.chunksize)
	if size%s.chunksize != 0 {
		numChunks += 1
	}

	s.log.Infow(
		"Uploading file using multi-part file",
		"key", file.Key,
		"num_parts", numChunks,
	)

	uploadID, err := s.createMultiPartUpload(ctx, file)
	if err != nil {
		return "", "", err
	}

	uploadOutputs := make([]*UploadPartOutput, numChunks)

	for i := 0; i < numChunks; i++ {
		partNum := int32(i + 1)
		s.log.Debugw(
			"Upload part start",
			"key", file.Key,
			"part", partNum,
		)
		uploadOutput, err := s.uploadPart(ctx, uploadID, partNum, size, file)
		if err != nil {
			return "", "", fmt.Errorf("unable to file part: %w", err)
		}
		s.log.Debugw(
			"Upload part finish",
			"key", file.Key,
			"part", partNum,
		)
		uploadOutputs[i] = uploadOutput
	}

	return s.closeMultiPartUpload(ctx, uploadID, uploadOutputs, file)
}

func (s *Store) createMultiPartUpload(
	ctx context.Context,
	file *File,
) (string, error) {

	input := file.ToCreateMultipartUploadInput()

	output, err := s.client.CreateMultipartUpload(ctx, (*s3.CreateMultipartUploadInput)(input))
	if err != nil {
		return "", fmt.Errorf("unable to create multipart file: %w", err)
	}

	return *output.UploadId, nil
}

func (s *Store) uploadPart(
	ctx context.Context,
	uploadID string,
	partNum int32,
	size int64,
	file *File,
) (*UploadPartOutput, error) {

	chunksize := s.chunksize
	remainingBytes := size - int64(partNum-1)*(s.chunksize)
	if remainingBytes < chunksize {
		chunksize = remainingBytes
	}

	input := file.ToUploadPartInput(uploadID, partNum, chunksize)

	output, err := s.client.UploadPart(ctx, (*s3.UploadPartInput)(input))
	if err != nil {
		return nil, fmt.Errorf("unable to file multipart part: %w", err)
	}

	return (*UploadPartOutput)(output), nil
}

func (s *Store) closeMultiPartUpload(
	ctx context.Context,
	uploadID string,
	parts []*UploadPartOutput,
	file *File,
) (string, string, error) {

	input := file.ToCompleteMultipartUploadInput(uploadID, parts)

	output, err := s.client.CompleteMultipartUpload(ctx, (*s3.CompleteMultipartUploadInput)(input))
	if err != nil {
		return "", "", err
	}

	version := ""
	if output.VersionId != nil {
		version = *output.VersionId
	}

	return *output.ETag, version, nil
}

func (s *Store) reportElapsedFileUploadTime(start time.Time, fileUpload *File) {
	elapsed := time.Since(start)
	s.log.Infow(
		"Upload to store finished",
		"key", fileUpload.Key,
		"bucket", fileUpload.Bucket,
		"elapsed_time", elapsed,
	)
}
