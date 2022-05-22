package dirscanner

import (
	"context"
	"encoding/base64"
	"io/fs"
	"sync"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/dirscanner.go -package=mocks -source=$GOFILE

// FileUploadHandler is the interface required to handle file uploads produced
// by the DirScanner instance.
type FileUploadHandler interface {
	HandleFileUpload(
		ctx context.Context,
		upload *models.FileUpload,
	) (*models.FileUpload, error)
}

// VersionCalculator is the interface required to derive a version string from a
// file path.
type VersionCalculator interface {
	CalculateVersion(path string) (string, error)
}

// Salter is the interface required to derive a cryptographic salt from a file
// path.
type Salter interface {
	Salt(path string) ([]byte, error)
}

// DirScanner encapsulates the logic for scanning a directory hierarchy and
// generating FileUpload objects from the files within.
type DirScanner struct {
	fs                fs.FS
	vc                VersionCalculator
	salter            Salter
	bucket            string
	encAlg            models.EncryptionAlgorithm
	numHandlerThreads int
	uploadHandlers    []FileUploadHandler
	uploadQueue       chan *models.FileUpload
	wg                *sync.WaitGroup
	log               *zap.SugaredLogger
}

// Scan traverses the filesystem and runs all registered uploadHandlers on all
// regular files.
func (s *DirScanner) Scan(ctx context.Context) error {
	s.uploadQueue = make(chan *models.FileUpload)

	for i := 0; i < s.numHandlerThreads; i++ {
		s.wg.Add(1)
		go s.handleFileUploads(ctx)
	}

	err := fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		s.log.Debugw("Handling path", "path", path)
		if err != nil {
			s.log.Warnw("Skipping file due to error", "error", err, "path", path)
			return nil
		}

		if !d.Type().IsRegular() {
			s.log.Infow("Skipping irregular file type", "path", path, "type", d.Type())
			return nil
		}

		version, err := s.vc.CalculateVersion(path)
		if err != nil {
			s.log.Warnw("Failed to calculate file version", "error", err, "path", path)
			return nil
		}
		salt, err := s.salter.Salt(path)
		if err != nil {
			s.log.Warnw("Failed to generate file salt", "error", err, "path", path)
			return nil
		}

		fileUpload := &models.FileUpload{
			LocalPath:           path,
			Bucket:              s.bucket,
			Version:             version,
			Salt:                base64.RawStdEncoding.EncodeToString(salt),
			EncryptionAlgorithm: s.encAlg,
		}

		s.uploadQueue <- fileUpload

		return nil
	})

	close(s.uploadQueue)

	s.wg.Wait()

	return err
}

func (s *DirScanner) handleFileUploads(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case fu, ok := <-s.uploadQueue:
			if !ok {
				return
			}
			for _, handler := range s.uploadHandlers {
				_, err := handler.HandleFileUpload(ctx, fu)
				if err != nil {
					s.log.Warnw("Error handling file upload", "error", err, "file_upload", fu)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
