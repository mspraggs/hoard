package store

import (
	"time"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/store/models"
)

func reportElapsedFileUploadTime(
	log *zap.SugaredLogger,
	start time.Time,
	fileUpload *models.FileUpload,
) {

	elapsed := time.Since(start)
	log.Infow(
		"Upload to store finished",
		"key", fileUpload.Key,
		"bucket", fileUpload.Bucket,
		"elapsed_time", elapsed,
	)
}
