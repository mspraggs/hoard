package uploader

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/uploader.go -package=mocks -source=$GOFILE

type Checksummer interface {
	Checksum(reader io.Reader) (models.Checksum, error)
}

type Uploader interface {
	Upload(
		ctx context.Context,
		file io.Reader,
		csc Checksummer,
		upload *fsmodels.FileUpload,
	) error
}

type UploaderSelector struct {
	sizeThreshold     int64
	smallFileUploader Uploader
	largeFileUploader Uploader
}

func NewUploaderSelector(
	smallFileUploader,
	largeFileUploader Uploader,
	threshold int64,
) *UploaderSelector {

	return &UploaderSelector{threshold, smallFileUploader, largeFileUploader}
}

func (us *UploaderSelector) SelectUploader(file fs.File) (Uploader, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to look up file info: %w", err)
	}
	if stat.Size() > us.sizeThreshold {
		return us.largeFileUploader, nil
	}
	return us.smallFileUploader, nil
}
