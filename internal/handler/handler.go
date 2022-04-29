package handler

import (
	"context"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/handler.go -package=mocks -source=$GOFILE

type FileRegistry interface {
	RegisterFileUpload(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	MarkFileUploadUploaded(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
}

type FileStore interface {
	StoreFileUpload(ctx context.Context, FileUpload *models.FileUpload) (*models.FileUpload, error)
}

type Handler struct {
	fs   FileStore
	freg FileRegistry
}

func New(fs FileStore, freg FileRegistry) *Handler {
	return &Handler{fs, freg}
}

func (h *Handler) HandleFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	createdFileUpload, err := h.freg.RegisterFileUpload(ctx, fileUpload)
	if err != nil {
		return nil, fmt.Errorf("error creating file upload: %w", err)
	}

	if !createdFileUpload.UploadedAtTimestamp.IsZero() {
		// TODO: Logging
		return createdFileUpload, nil
	}

	uploadedFileUpload, err := h.fs.StoreFileUpload(ctx, createdFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error while uploading file to file store: %w", err)
	}

	uploadedAndMarkedFileUpload, err := h.freg.MarkFileUploadUploaded(ctx, uploadedFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error marking file upload as uploaded: %w", err)
	}

	return uploadedAndMarkedFileUpload, nil
}
