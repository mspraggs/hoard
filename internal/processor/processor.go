package processor

import (
	"context"
	"io/fs"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/processor.go -package=mocks -source=$GOFILE

// Checksum defines a CRC32 checksum as an unsigned 32-bit integer.
type Checksum uint32

// File encapsulates all information associated with a file.
type File struct {
	Key       string
	LocalPath string
	Checksum  Checksum
	Bucket    string
	ETag      string
	Version   string
}

// KeyGenerator defines the interface required to generate a random key.
type KeyGenerator interface {
	GenerateKey() string
}

// Registry specifies the interface required to register and update the
// registry of uploaded files.
type Registry interface {
	Create(ctx context.Context, file *File) (*File, error)
	FetchLatest(ctx context.Context, path string) (*File, error)
}

// Uploader specifies the interface required to upload files.
type Uploader interface {
	Upload(ctx context.Context, file *File) (*File, error)
}

// Processor encapsulates the logic required to create a file and store it in
// the file store.
type Processor struct {
	log      *zap.SugaredLogger
	fs       fs.FS
	keyGen   KeyGenerator
	registry Registry
	uploader Uploader
}

// New instantiates a new Processor instance with provided file store and
// registry.
func New(fs fs.FS, keyGen KeyGenerator, uploader Uploader, registry Registry) *Processor {
	log := util.MustNewLogger()
	return &Processor{
		log:      log,
		fs:       fs,
		keyGen:   keyGen,
		registry: registry,
		uploader: uploader,
	}
}
