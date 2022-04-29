package util

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"

	"github.com/mspraggs/hoard/internal/models"
)

type EncryptionKeyGenerator struct {
	secret    []byte
	algorithm models.EncryptionAlgorithm
	time      uint32
	memory    uint32
	threads   uint8
	keyLen    uint32
}

func NewEncryptionKeyGenerator(
	secret []byte,
	algorithm models.EncryptionAlgorithm,
	time uint32,
	memory uint32,
	threads uint8,
	keyLen uint32,
) *EncryptionKeyGenerator {

	return &EncryptionKeyGenerator{secret, algorithm, time, memory, threads, keyLen}
}

func (ekg *EncryptionKeyGenerator) GenerateKey(
	fileUpload *models.FileUpload,
) (models.EncryptionKey, error) {

	keyLen, err := fileUpload.EncryptionAlgorithm.KeySize()
	if err != nil {
		return models.EncryptionKey(""), fmt.Errorf("unable to generate encryption key: %w", err)
	}
	keyBytes := argon2.Key(
		ekg.secret,
		fileUpload.Salt,
		ekg.time,
		ekg.memory,
		ekg.threads,
		keyLen,
	)

	return models.EncryptionKey(base64.StdEncoding.EncodeToString(keyBytes)), nil
}
