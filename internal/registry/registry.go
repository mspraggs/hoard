package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/registry.go -package=mocks -source=$GOFILE

// Clock defines the interface required to fetch the current time.
type Clock interface {
	Now() time.Time
}

// Store defines the interface required to interact with the file registry
// storage backend.
type Store interface {
	GetFileUploadByChangeRequestID(
		ctx context.Context,
		requestID string,
	) (*models.FileUpload, models.ChangeType, error)
	InsertFileUpload(
		ctx context.Context,
		requestID string,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	UpdateFileUpload(
		ctx context.Context,
		requestID string,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
}

// TxnFunc defines the closure expected by the InTransactioner interface when
// handling a database transaction.
type TxnFunc func(context.Context, Store) (interface{}, error)

// InTransactioner defines the interface required to interact with a database
// within a transaction.
type InTransactioner interface {
	InTransaction(ctx context.Context, fn TxnFunc) (interface{}, error)
}

// RequestIDMaker defines the interface required to construct a request ID for a
// given file upload.
type RequestIDMaker interface {
	MakeRequestID(fileUpload *models.FileUpload) (string, error)
}

// Registry encapsulates the logic required to interact with a register of
// file uploads. The registry maintains a record of details associated with a
// file upload, including a file version string and the timestamp at which the
// file was uploaded.
type Registry struct {
	clock          Clock
	inTxner        InTransactioner
	requestIDMaker RequestIDMaker
}

// New instantiates a new Registry using the provided Clock, InTransactioner
// and RequestIDMaker instances.
func New(clock Clock, inTxner InTransactioner, requestIDMaker RequestIDMaker) *Registry {
	return &Registry{clock, inTxner, requestIDMaker}
}

// RegisterFileUpload creates the file upload in the registry storage backend
// using an idempotency key derived from the local file path and the file
// version string.
func (r *Registry) RegisterFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	requestID, err := r.requestIDMaker.MakeRequestID(fileUpload)
	if err != nil {
		return nil, fmt.Errorf("unable to make request ID: %w", err)
	}

	opaqueCreatedFileUpload, err := r.inTxner.InTransaction(
		ctx,
		func(c context.Context, s Store) (interface{}, error) {
			fileUpload.CreatedAtTimestamp = r.clock.Now()
			existingFileUpload, changeType, err := s.GetFileUploadByChangeRequestID(c, requestID)

			if err == nil {
				if changeType == models.ChangeTypeCreate {
					return existingFileUpload, nil
				}
				return nil, pkgerrors.ErrInvalidRequestID
			}

			if !errors.Is(err, pkgerrors.ErrNotFound) {
				return nil, err
			}

			return s.InsertFileUpload(c, requestID, fileUpload)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while inserting file upload into db: %w", err)
	}

	createdFileUpload := opaqueCreatedFileUpload.(*models.FileUpload)

	return createdFileUpload, nil
}

// GetUploadedFileUpload fetches an existing file upload from the database using
// the provided ID.
func (r *Registry) GetUploadedFileUpload(
	ctx context.Context,
	ID string,
) (*models.FileUpload, error) {

	opaqueUploadedFileUpload, err := r.inTxner.InTransaction(
		ctx,
		func(c context.Context, s Store) (interface{}, error) {
			existingFileUpload, changeType, err := s.GetFileUploadByChangeRequestID(c, ID)

			if err != nil {
				if !errors.Is(err, pkgerrors.ErrNotFound) {
					return nil, err
				} else {
					return nil, nil
				}
			}

			if changeType == models.ChangeTypeUpdate && existingFileUpload.IsUploaded() {
				return existingFileUpload, nil
			}
			return nil, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving file upload from db: %w", err)
	}
	if opaqueUploadedFileUpload == nil {
		return nil, nil
	}

	uploadedFileUpload := opaqueUploadedFileUpload.(*models.FileUpload)

	return uploadedFileUpload, nil
}

// MarkFileUploadUploaded marks a file upload as uploaded using the store held
// by the file registry.
func (r *Registry) MarkFileUploadUploaded(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	markUploaded := func() *models.FileUpload {
		fileUpload.UploadedAtTimestamp = r.clock.Now()
		return fileUpload
	}
	return r.updateFileUpload(
		ctx, fileUpload.ID+"_uploaded", models.ChangeTypeUpdate, markUploaded,
	)
}

// MarkFileUploadDeleted marks a file upload as uploaded using the store held
// by the file registry.
func (r *Registry) MarkFileUploadDeleted(
	ctx context.Context,
	fileUpload *models.FileUpload,
) error {

	markDeleted := func() *models.FileUpload {
		fileUpload.DeletedAtTimestamp = r.clock.Now()
		return fileUpload
	}
	if _, err := r.updateFileUpload(
		ctx, fileUpload.ID+"_deleted", models.ChangeTypeDelete, markDeleted,
	); err != nil {
		return err
	}
	return nil
}

func (r *Registry) updateFileUpload(
	ctx context.Context,
	requestID string,
	changeType models.ChangeType,
	getMutatedFileUpload func() *models.FileUpload,
) (*models.FileUpload, error) {

	result, err := r.inTxner.InTransaction(
		ctx,
		func(c context.Context, s Store) (interface{}, error) {
			existingFileUpload, existingChangeType, err :=
				s.GetFileUploadByChangeRequestID(c, requestID)

			if err == nil {
				if existingChangeType == changeType {
					return existingFileUpload, nil
				}
				return nil, pkgerrors.ErrInvalidRequestID
			}

			if !errors.Is(err, pkgerrors.ErrNotFound) {
				return nil, err
			}

			fileUpload := getMutatedFileUpload()
			return s.UpdateFileUpload(c, requestID, fileUpload)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while updating file upload in db: %w", err)
	}

	return result.(*models.FileUpload), nil
}
