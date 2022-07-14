package util_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/util"
)

type EncryptionKeyGeneratorTestSuite struct {
	suite.Suite
}

func TestEncryptionKeyGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(EncryptionKeyGeneratorTestSuite))
}

func (s *EncryptionKeyGeneratorTestSuite) TestGenerateKey() {
	secret := []byte{1, 2, 3}
	salt := []byte{4, 5, 6}
	keyLen := uint32(32)

	s.Run("generates and returns key for file upload", func() {
		expectedKey := []byte{
			0x2b, 0xc1, 0x2d, 0xc, 0x43, 0x8d, 0xb5, 0x3c, 0x75, 0x86, 0xee,
			0x84, 0x9c, 0xc6, 0x28, 0x68, 0x3d, 0x10, 0xe9, 0xdb, 0x6, 0x84,
			0x59, 0xd5, 0xef, 0x9a, 0xa1, 0x34, 0x94, 0x97, 0xfd, 0xd2,
		}

		encKeyGen := util.NewEncryptionKeyGenerator(secret, keyLen)

		key := encKeyGen.GenerateKey(salt)

		s.Equal(expectedKey, key)
	})
}
