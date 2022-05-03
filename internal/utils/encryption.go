package utils

import (
	"fmt"

	"golang.org/x/crypto/argon2"

	"github.com/mspraggs/hoard/internal/models"
)

const (
	memory  uint32 = 1024 * 64
	time    uint32 = 1
	threads uint8  = 1
)

// EncryptionKeyGenerator provides a way to generate encryption keys given some
// secret and file upload instance. The generator uses the Argon2 algorithm to
// turn the provided secret and salt into an encryption key.
type EncryptionKeyGenerator struct {
	secret []byte
}

// NewEncryptionKeyGenerator instantiates a new EncryptionKeyGenerator instance
// with the provided secret.
func NewEncryptionKeyGenerator(secret []byte) *EncryptionKeyGenerator {
	return &EncryptionKeyGenerator{secret}
}

// GenerateKey generates an encryption key from the provided file upload. The
// key size is determined based on the encryption algorithm attached to the file
// upload.
func (ekg *EncryptionKeyGenerator) GenerateKey(
	fileUpload *models.FileUpload,
) (models.EncryptionKey, error) {

	keyLen, err := fileUpload.EncryptionAlgorithm.KeySize()
	if err != nil {
		return models.EncryptionKey(nil), fmt.Errorf("unable to generate encryption key: %w", err)
	}
	keyBytes := argon2.Key(
		ekg.secret,
		fileUpload.Salt,
		time,
		memory,
		threads,
		keyLen,
	)

	return models.EncryptionKey(keyBytes), nil
}
