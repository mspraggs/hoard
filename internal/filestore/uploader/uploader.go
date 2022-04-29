package uploader

import (
	"io"

	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/uploader.go -package=mocks -source=$GOFILE

type Checksummer interface {
	Checksum(reader io.Reader) (models.Checksum, error)
}
