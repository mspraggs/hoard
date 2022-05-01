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

type EncryptionKeyGenerator struct {
	secret []byte
}

func NewEncryptionKeyGenerator(secret []byte) *EncryptionKeyGenerator {

	return &EncryptionKeyGenerator{secret}
}

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
