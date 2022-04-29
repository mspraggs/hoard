package utils_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/utils"
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
	fileUpload := &models.FileUpload{
		Salt:                salt,
		EncryptionAlgorithm: models.EncryptionAlgorithmAES256,
	}

	s.Run("generates and returns key for file upload", func() {
		expectedKey := models.EncryptionKey([]byte{
			236, 53, 40, 73, 35, 185, 53, 111, 126, 153, 77, 187, 244, 27, 220, 145, 22, 226, 7, 11,
			79, 28, 225, 79, 163, 86, 200, 229, 42, 226, 44, 3, 138, 243, 101, 5, 62, 249, 229, 90,
			7, 51, 250, 236, 163, 255, 222, 120, 60, 241, 6, 35, 224, 62, 229, 99, 146, 194, 70, 65,
			48, 221, 200, 77, 119, 107, 247, 228, 69, 184, 30, 112, 220, 100, 245, 4, 125, 47, 123,
			98, 58, 118, 167, 100, 87, 195, 136, 12, 0, 58, 69, 150, 93, 16, 203, 228, 116, 227,
			139, 204, 252, 9, 159, 79, 91, 154, 204, 5, 177, 2, 43, 59, 169, 211, 180, 203, 166,
			163, 225, 80, 76, 32, 207, 123, 151, 106, 62, 243, 230, 246, 91, 194, 14, 69, 92, 228,
			105, 160, 211, 142, 194, 98, 109, 146, 222, 71, 27, 225, 237, 166, 135, 195, 58, 120,
			59, 83, 141, 233, 201, 24, 162, 80, 251, 116, 79, 117, 138, 72, 99, 194, 175, 146, 60,
			57, 169, 212, 172, 69, 151, 214, 81, 12, 254, 254, 90, 130, 107, 98, 15, 17, 8, 159,
			175, 145, 140, 129, 188, 134, 182, 77, 27, 231, 8, 99, 184, 193, 76, 19, 57, 191, 56,
			225, 72, 95, 110, 208, 169, 99, 159, 155, 84, 164, 28, 89, 101, 168, 3, 84, 96, 162,
			248, 185, 30, 95, 150, 4, 99, 110, 19, 127, 12, 167, 6, 177, 15, 175, 96, 46, 236, 105,
			175, 87, 107, 151, 111, 105,
		})

		encKeyGen := utils.NewEncryptionKeyGenerator(secret)

		key, err := encKeyGen.GenerateKey(fileUpload)

		s.Require().NoError(err)
		s.Equal(expectedKey, key)
	})

	s.Run("returns error for unsupported encryption algorithm", func() {
		expectedKey := models.EncryptionKey(nil)

		fileUpload := &models.FileUpload{
			Salt:                salt,
			EncryptionAlgorithm: models.EncryptionAlgorithm(0),
		}

		encKeyGen := utils.NewEncryptionKeyGenerator(nil)

		key, err := encKeyGen.GenerateKey(fileUpload)

		s.Equal(expectedKey, key)
		s.ErrorContains(err, "unable to generate")
	})
}
