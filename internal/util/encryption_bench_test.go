package util_test

import (
	"crypto/rand"
	"testing"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/util"
)

const (
	numSecrets = 100000
	secretLen  = 50
)

func BenchmarkGenerateKey(b *testing.B) {
	salts, err := generateBenchmarkSalts(numSecrets, secretLen)
	if err != nil {
		b.Errorf("Unable to generate salts: %v", err)
	}

	fileUploads := make([]*models.FileUpload, len(salts))
	for i, salt := range salts {
		fileUploads[i] = &models.FileUpload{
			Salt:                salt,
			EncryptionAlgorithm: models.EncryptionAlgorithmAES256,
		}
	}

	keyGen := util.NewEncryptionKeyGenerator([]byte("somerandompassword"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		keyGen.GenerateKey(fileUploads[i%len(fileUploads)])
	}
}

func generateBenchmarkSalts(num, len int) ([][]byte, error) {
	salts := make([][]byte, numSecrets)

	for i := 0; i < num; i++ {
		salt := make([]byte, secretLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		salts[i] = salt
	}

	return salts, nil
}
