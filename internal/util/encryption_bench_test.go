package util_test

import (
	"crypto/rand"
	"testing"

	"github.com/mspraggs/hoard/internal/util"
)

const (
	numSecrets = 100000
	keyLen     = 32
	secretLen  = 50
)

func BenchmarkGenerateKey(b *testing.B) {
	salts, err := generateBenchmarkSalts(numSecrets, secretLen)
	if err != nil {
		b.Errorf("Unable to generate salts: %v", err)
	}

	keyGen := util.NewEncryptionKeyGenerator([]byte("somerandompassword"), keyLen)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		keyGen.GenerateKey(salts[i%len(salts)])
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
