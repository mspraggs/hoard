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
		expectedKey := models.EncryptionKey(
			"7DUoSSO5NW9+mU279BvckRbiBwtPHOFPo1bI5SriLAOK82UFPvnlWgcz+uyj/954PPEGI+A+5WOSwkZBMN3I" +
				"TXdr9+RFuB5w3GT1BH0ve2I6dqdkV8OIDAA6RZZdEMvkdOOLzPwJn09bmswFsQIrO6nTtMumo+FQTCDP" +
				"e5dqPvPm9lvCDkVc5Gmg047CYm2S3kcb4e2mh8M6eDtTjenJGKJQ+3RPdYpIY8Kvkjw5qdSsRZfWUQz+" +
				"/lqCa2IPEQifr5GMgbyGtk0b5whjuMFMEzm/OOFIX27QqWOfm1SkHFllqANUYKL4uR5flgRjbhN/DKcG" +
				"sQ+vYC7saa9Xa5dvaQ==",
		)

		encKeyGen := utils.NewEncryptionKeyGenerator(secret)

		key, err := encKeyGen.GenerateKey(fileUpload)

		s.Require().NoError(err)
		s.Equal(expectedKey, key)
	})

	s.Run("returns error for unsupported encryption algorithm", func() {
		expectedKey := models.EncryptionKey("")

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
