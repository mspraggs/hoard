package util_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/models"
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
	salt := "BAUG"
	fileUpload := &models.FileUpload{
		Salt:                salt,
		EncryptionAlgorithm: models.EncryptionAlgorithmAES256,
	}

	s.Run("generates and returns key for file upload", func() {
		expectedKey := models.EncryptionKey([]byte{
			0x2b, 0xc1, 0x2d, 0xc, 0x43, 0x8d, 0xb5, 0x3c, 0x75, 0x86, 0xee,
			0x84, 0x9c, 0xc6, 0x28, 0x68, 0x3d, 0x10, 0xe9, 0xdb, 0x6, 0x84,
			0x59, 0xd5, 0xef, 0x9a, 0xa1, 0x34, 0x94, 0x97, 0xfd, 0xd2,
		})

		encKeyGen := util.NewEncryptionKeyGenerator(secret)

		key, err := encKeyGen.GenerateKey(fileUpload)

		s.Require().NoError(err)
		s.Equal(expectedKey, key)
	})

	s.Run("returns error", func() {
		s.Run("for unsupported encryption algorithm", func() {
			expectedKey := models.EncryptionKey(nil)

			fileUpload := &models.FileUpload{
				Salt:                salt,
				EncryptionAlgorithm: models.EncryptionAlgorithm(0),
			}

			encKeyGen := util.NewEncryptionKeyGenerator(nil)

			key, err := encKeyGen.GenerateKey(fileUpload)

			s.Equal(expectedKey, key)
			s.ErrorContains(err, "unable to generate")
		})
		s.Run("for invalid salt", func() {
			expectedKey := models.EncryptionKey(nil)

			fileUpload := &models.FileUpload{
				Salt:                "&*(",
				EncryptionAlgorithm: models.EncryptionAlgorithmAES256,
			}

			encKeyGen := util.NewEncryptionKeyGenerator(nil)

			key, err := encKeyGen.GenerateKey(fileUpload)

			s.Equal(expectedKey, key)
			s.ErrorContains(err, "unable to decode")
		})
	})

}
