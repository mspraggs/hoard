package store

import (
	"context"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
	fsmodels "github.com/mspraggs/hoard/internal/store/models"
)

// EraseFileUpload removes the provided file upload from the storage backend.
func (s *Store) EraseFileUpload(ctx context.Context, fileUpload *models.FileUpload) error {
	upload := fsmodels.NewFileUploadFromBusiness(
		fileUpload.EncryptionAlgorithm, models.EncryptionKey{}, s.csAlg, s.sc, 0, fileUpload, nil,
	)

	if err := s.client.Delete(ctx, upload); err != nil {
		return fmt.Errorf("unable to delete file: %w", err)
	}

	return nil
}
