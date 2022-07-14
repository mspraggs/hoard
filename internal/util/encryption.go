package util

import (
	"encoding/json"

	"golang.org/x/crypto/argon2"
)

const (
	memParam  uint32 = 1024 * 64
	timeParam uint32 = 1
	threads   uint8  = 1
)

// EncryptionKeyGenerator provides a way to generate encryption keys given some
// secret and file upload instance. The generator uses the Argon2 algorithm to
// turn the provided secret and salt into an encryption key.
type EncryptionKeyGenerator struct {
	secret []byte
	keyLen uint32
}

// NewEncryptionKeyGenerator instantiates a new EncryptionKeyGenerator instance
// with the provided secret and key length.
func NewEncryptionKeyGenerator(secret []byte, keyLen uint32) *EncryptionKeyGenerator {
	return &EncryptionKeyGenerator{secret, keyLen}
}

// GenerateKey generates an encryption key from the provided salt and key
// length.
func (ekg *EncryptionKeyGenerator) GenerateKey(salt []byte) []byte {
	return argon2.Key(
		ekg.secret,
		salt,
		timeParam,
		memParam,
		threads,
		ekg.keyLen,
	)
}

// String serialises the parameters used by the encryption key generator.
func (ekg *EncryptionKeyGenerator) String() string {
	params := map[string]interface{}{
		"algorithm": "argon2",
		"time":      timeParam,
		"memory":    memParam,
		"threads":   threads,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	return string(paramsJSON)
}
