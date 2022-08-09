package processor

import (
	"context"
	"io/fs"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
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
	CTime     time.Time
	Bucket    string
	ETag      string
	Version   string
}

// KeyGenerator defines the interface required to generate a random key.
type KeyGenerator interface {
	GenerateKey() string
}

// CTimeGetter defines the interface required to get the change time from a
// FileInfo object.
type CTimeGetter interface {
	GetCTime(fi fs.File) (time.Time, error)
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

// Option is the type used to implement the functional options pattern for the
// Processor type.
type Option func(*Processor)

// Processor encapsulates the logic required to create a file and store it in
// the file store.
type Processor struct {
	log      *zap.SugaredLogger
	fs       fs.FS
	keyGen   KeyGenerator
	ctg      CTimeGetter
	registry Registry
	uploader Uploader
}

// New instantiates a new Processor instance with provided file store and
// registry.
func New(fs fs.FS, uploader Uploader, registry Registry, opts ...Option) *Processor {
	log := util.MustNewLogger()
	p := &Processor{
		log:      log,
		fs:       fs,
		keyGen:   keyGen{},
		ctg:      ctimeGetter{},
		registry: registry,
		uploader: uploader,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithKeyGenerator returns an option for setting the way in which a Processor
// generates a key for a new file.
func WithKeyGenerator(kg KeyGenerator) Option {
	return func(p *Processor) {
		p.keyGen = kg
	}
}

// WithCTimeGetter returns an option for setting the way in which a Processor
// gets the ctime from a FileInfo object.
func WithCTimeGetter(ctg CTimeGetter) Option {
	return func(p *Processor) {
		p.ctg = ctg
	}
}

type keyGen struct{}

// GenerateKey returns a random UUID in accordance with the KeyGenerator
// interface.
func (g keyGen) GenerateKey() string {
	return uuid.NewString()
}

type ctimeGetter struct{}

// GetCTime extracts the ctime from the provided file object in accordance with
// the CTimeGetter interface.
func (ctg ctimeGetter) GetCTime(f fs.File) (time.Time, error) {
	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}
	ctime := fi.Sys().(*syscall.Stat_t).Ctim
	return time.Unix(int64(ctime.Sec), int64(ctime.Nsec)).UTC().Truncate(time.Microsecond), nil
}
