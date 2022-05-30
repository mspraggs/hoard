package processor

import (
	"context"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/processor.go -package=mocks -source=$GOFILE

// Registry specifies the interface required to register and update the
// registry of uploaded files.
type Registry interface {
	RegisterFileUpload(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	GetUploadedFileUpload(ctx context.Context, ID string) (*models.FileUpload, error)
	MarkFileUploadUploaded(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	MarkFileUploadDeleted(ctx context.Context, fileUpload *models.FileUpload) error
}

// Store specifies the interface required to upload files.
type Store interface {
	EraseFileUpload(ctx context.Context, FileUpload *models.FileUpload) error
	StoreFileUpload(ctx context.Context, FileUpload *models.FileUpload) (*models.FileUpload, error)
}

// Processor encapsulates the logic required to register a file upload and store
// it in the file store.
type Processor struct {
	fs   Store
	freg Registry
	log  *zap.SugaredLogger
}

// New instantiates a new Processor instance with provided file store and
// registry.
func New(fs Store, freg Registry) *Processor {
	log := util.MustNewLogger()
	return &Processor{fs, freg, log}
}
