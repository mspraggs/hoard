package db

import (
	"context"

	"github.com/doug-martin/goqu"

	dbmodels "github.com/mspraggs/hoard/internal/db/models"
	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/models"
)

// Store encapsulates the logic for interacting with a database transaction.
type Store struct {
	tx *goqu.TxDatabase
}

// NewStore instantiates a new Store instance with the provided database
// transaction.
func NewStore(tx *goqu.TxDatabase) *Store {
	return &Store{tx}
}

// GetFileUploadByChangeRequestID fetches a file upload and change type from the
// database using the provided request ID.
func (s *Store) GetFileUploadByChangeRequestID(
	ctx context.Context,
	requestID string,
) (*models.FileUpload, models.ChangeType, error) {

	fileUploadHistoryRow := &dbmodels.FileUploadHistoryRow{}
	found, err := s.tx.
		From("file_uploads_history").
		Select(goqu.Star()).
		Where(goqu.I("request_id").Eq(requestID)).
		ScanStructContext(ctx, fileUploadHistoryRow)
	if err != nil {
		return nil, models.ChangeType(0), err
	}
	if !found {
		return nil, models.ChangeType(0), pkgerrors.ErrNotFound
	}

	fileUpload, _, changeType := fileUploadHistoryRow.ToBusiness()
	return fileUpload, changeType, nil
}

// InsertFileUpload inserts a new file upload into the database. A row is also
// added to the append-only history table for this resource.
func (s *Store) InsertFileUpload(
	ctx context.Context,
	requestID string,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	fileUploadHistoryRow := dbmodels.NewFileUploadHistoryRowFromBusiness(
		requestID, models.ChangeTypeCreate, fileUpload,
	)
	if err := s.insertFileUploadHistoryRow(ctx, fileUploadHistoryRow); err != nil {
		return nil, err
	}

	fileUploadRow := dbmodels.NewFileUploadRowFromBusiness(fileUpload)

	_, err := s.tx.From("file_uploads").Insert(fileUploadRow).ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	return fileUpload, nil
}

// UpdateFileUpload updates an existing file upload in the database. A row is
// also added to the append-only history table for this resource.
func (s *Store) UpdateFileUpload(
	ctx context.Context,
	requestID string,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	fileUploadHistoryRow := dbmodels.NewFileUploadHistoryRowFromBusiness(
		requestID, models.ChangeTypeUpdate, fileUpload,
	)
	if err := s.insertFileUploadHistoryRow(ctx, fileUploadHistoryRow); err != nil {
		return nil, err
	}

	fileUploadRow := dbmodels.NewFileUploadRowFromBusiness(fileUpload)

	_, err := s.tx.
		From("file_uploads").
		Where(goqu.I("id").Eq(fileUpload.ID)).
		Update(fileUploadRow).
		ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	return fileUpload, nil
}

func (s *Store) insertFileUploadHistoryRow(
	ctx context.Context,
	fileUploadHistoryRow *dbmodels.FileUploadHistoryRow,
) error {

	_, err := s.tx.
		From("file_uploads_history").
		Insert(fileUploadHistoryRow).
		ExecContext(ctx)
	return err
}
