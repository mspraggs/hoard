package models

import (
	"time"

	"github.com/mspraggs/hoard/internal/models"
)

type ChangeType int

const (
	ChangeTypeCreate ChangeType = 1
	ChangeTypeUpdate ChangeType = 2
)

type FileUploadRow struct {
	ID                  string    `db:"id"`
	LocalPath           string    `db:"local_path"`
	Bucket              string    `db:"bucket"`
	Version             string    `db:"version"`
	Salt                []byte    `db:"salt"`
	CreatedAtTimestamp  time.Time `db:"created_at_timestamp"`
	UploadedAtTimestamp time.Time `db:"uploaded_at_timestamp"`
}

type FileUploadHistoryRow struct {
	RequestID           string     `db:"request_id"`
	ID                  string     `db:"id"`
	LocalPath           string     `db:"local_path"`
	Bucket              string     `db:"bucket"`
	Version             string     `db:"version"`
	Salt                []byte     `db:"salt"`
	CreatedAtTimestamp  time.Time  `db:"created_at_timestamp"`
	UploadedAtTimestamp time.Time  `db:"uploaded_at_timestamp"`
	ChangeType          ChangeType `db:"change_type"`
}

func NewChangeTypeFromBusiness(c models.ChangeType) ChangeType {
	switch c {
	case models.ChangeTypeCreate:
		return ChangeTypeCreate
	case models.ChangeTypeUpdate:
		return ChangeTypeUpdate
	default:
		return ChangeType(0)
	}
}

func (c ChangeType) ToBusiness() models.ChangeType {
	switch c {
	case ChangeTypeCreate:
		return models.ChangeTypeCreate
	case ChangeTypeUpdate:
		return models.ChangeTypeUpdate
	default:
		return models.ChangeType(0)
	}
}

func NewFileUploadRowFromBusiness(from *models.FileUpload) *FileUploadRow {
	return &FileUploadRow{
		ID:                  from.ID,
		LocalPath:           from.LocalPath,
		Bucket:              from.LocalPath,
		Version:             from.Version,
		Salt:                from.Salt,
		CreatedAtTimestamp:  from.CreatedAtTimestamp,
		UploadedAtTimestamp: from.UploadedAtTimestamp,
	}
}

func (fu *FileUploadRow) ToBusiness() *models.FileUpload {
	return &models.FileUpload{
		ID:                  fu.ID,
		LocalPath:           fu.LocalPath,
		Bucket:              fu.Bucket,
		Version:             fu.Version,
		Salt:                fu.Salt,
		CreatedAtTimestamp:  fu.CreatedAtTimestamp,
		UploadedAtTimestamp: fu.UploadedAtTimestamp,
	}
}

func NewFileUploadHistoryRowFromBusiness(
	requestID string,
	changeType models.ChangeType,
	upload *models.FileUpload,
) *FileUploadHistoryRow {

	return &FileUploadHistoryRow{
		RequestID:           requestID,
		ID:                  upload.ID,
		LocalPath:           upload.LocalPath,
		Bucket:              upload.LocalPath,
		Version:             upload.Version,
		Salt:                upload.Salt,
		CreatedAtTimestamp:  upload.CreatedAtTimestamp,
		UploadedAtTimestamp: upload.UploadedAtTimestamp,
		ChangeType:          NewChangeTypeFromBusiness(changeType),
	}
}

func (fu *FileUploadHistoryRow) ToBusiness() (*models.FileUpload, string, models.ChangeType) {
	return &models.FileUpload{
		ID:                  fu.ID,
		LocalPath:           fu.LocalPath,
		Bucket:              fu.Bucket,
		Version:             fu.Version,
		Salt:                fu.Salt,
		CreatedAtTimestamp:  fu.CreatedAtTimestamp,
		UploadedAtTimestamp: fu.UploadedAtTimestamp,
	}, fu.RequestID, fu.ChangeType.ToBusiness()
}
