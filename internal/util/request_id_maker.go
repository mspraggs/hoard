package util

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
)

// RequestIDMaker provides the logic for generating a request ID for a given
// file upload.
type RequestIDMaker struct{}

// NewRequestIDMaker instantiates a new RequestIDMaker instance using the
// provide filesystem.
func NewRequestIDMaker() *RequestIDMaker {
	return &RequestIDMaker{}
}

// MakeRequestID makes a unique creation request ID for the provided file
// upload by hashing the concatenated file path and version string.
func (m *RequestIDMaker) MakeRequestID(fileUpload *models.FileUpload) (string, error) {
	requestIdentifier := fmt.Sprintf("%v##%v", fileUpload.LocalPath, fileUpload.Version)
	requestIdentifierHash := md5.Sum([]byte(requestIdentifier))

	return base64.StdEncoding.EncodeToString(requestIdentifierHash[:]), nil
}
