package fileregistry

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/fileregistry.go -package=mocks -source=$GOFILE

type Clock interface {
	Now() time.Time
}

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

type TxnFunc func(context.Context, Store) (interface{}, error)

type InTransactioner interface {
	InTransaction(ctx context.Context, fn TxnFunc) (interface{}, error)
}

type FileRegistry struct {
	clock   Clock
	inTxner InTransactioner
}

func New(clock Clock, inTxner InTransactioner) *FileRegistry {
	return &FileRegistry{clock, inTxner}
}

func (r *FileRegistry) RegisterFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	hash := hashString(
		fileUpload.LocalPath + fileUpload.Version,
	)
	requestID := base64.StdEncoding.EncodeToString(hash)

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

func (r *FileRegistry) MarkFileUploadUploaded(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	opaqueUpdatedFileUpload, err := r.inTxner.InTransaction(
		ctx,
		func(c context.Context, s Store) (interface{}, error) {
			fileUpload.UploadedAtTimestamp = r.clock.Now()
			requestID := fileUpload.ID
			existingFileUpload, changeType, err := s.GetFileUploadByChangeRequestID(c, requestID)

			if err == nil {
				if changeType == models.ChangeTypeUpdate {
					return existingFileUpload, nil
				}
				return nil, pkgerrors.ErrInvalidRequestID
			}

			if !errors.Is(err, pkgerrors.ErrNotFound) {
				return nil, err
			}

			return s.UpdateFileUpload(c, requestID, fileUpload)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while updating file upload in db: %w", err)
	}

	return opaqueUpdatedFileUpload.(*models.FileUpload), nil
}

func hashString(s string) []byte {
	r := sha256.New()
	r.Write([]byte(s))
	return r.Sum(nil)
}
