package processor

import (
	"context"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
)

// UploadFileUpload registers the provided file upload in the file registry and
// uploads the file to the file store.
func (p *Processor) UploadFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	p.log.Debugw("Uploading file upload", "path", fileUpload.LocalPath)

	createdFileUpload, err := p.freg.RegisterFileUpload(ctx, fileUpload)
	if err != nil {
		return nil, fmt.Errorf("error creating file upload: %w", err)
	}
	p.log.Infow(
		"Registered file upload",
		"id", createdFileUpload.ID,
		"path", createdFileUpload.LocalPath,
	)

	uploadedFileUpload, err := p.freg.GetUploadedFileUpload(ctx, createdFileUpload.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving uploaded file upload: %w", err)
	}
	if uploadedFileUpload != nil {
		p.log.Infow(
			"Skipping uploaded file upload",
			"id", uploadedFileUpload.ID,
			"path", uploadedFileUpload.LocalPath,
		)
		return uploadedFileUpload, nil
	}

	uploadedFileUpload, err = p.fs.StoreFileUpload(ctx, createdFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error while uploading file to file store: %w", err)
	}
	p.log.Infow(
		"Stored file upload in storage backend",
		"id", uploadedFileUpload.ID,
		"path", uploadedFileUpload.LocalPath,
	)

	uploadedAndMarkedFileUpload, err := p.freg.MarkFileUploadUploaded(ctx, uploadedFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error marking file upload as uploaded: %w", err)
	}
	p.log.Infow(
		"Marked file upload as uploaded",
		"id", uploadedAndMarkedFileUpload.ID,
		"path", uploadedAndMarkedFileUpload.LocalPath,
	)

	return uploadedAndMarkedFileUpload, nil
}
