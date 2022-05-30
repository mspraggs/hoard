package processor

import (
	"context"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
)

// DeleteFileUpload deletes the provided file upload from the storage backend
// and marks it as deleted in the file registry.
func (p *Processor) DeleteFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) error {

	p.log.Debugw("Deleting file upload", "id", fileUpload.ID)

	if err := p.fs.EraseFileUpload(ctx, fileUpload); err != nil {
		return fmt.Errorf("error while erasing file from store: %w", err)
	}
	p.log.Infow(
		"Deleted file upload from storage backend",
		"id", fileUpload.ID,
		"path", fileUpload.LocalPath,
	)

	if err := p.freg.MarkFileUploadDeleted(ctx, fileUpload); err != nil {
		return fmt.Errorf("error marking file upload as uploaded: %w", err)
	}
	p.log.Infow(
		"Marked file upload as deleted",
		"id", fileUpload.ID,
		"path", fileUpload.LocalPath,
	)

	return nil
}
