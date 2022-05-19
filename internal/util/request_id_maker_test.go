package util_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/util"
)

type RequestIDMakerTestSuite struct {
	suite.Suite
}

func TestRequestIDMakerTestSuite(t *testing.T) {
	suite.Run(t, new(RequestIDMakerTestSuite))
}

func (s *RequestIDMakerTestSuite) TestMakeRequestID() {
	upload := &models.FileUpload{
		LocalPath: "/some/file/path",
		Version:   "latest-version-string",
	}

	s.Run("makes request ID for valid file", func() {
		expectedID := "YTHE+ckaAFnZKLfcy47dXw=="
		idMaker := util.NewRequestIDMaker()

		id, err := idMaker.MakeRequestID(upload)

		s.Require().NoError(err)
		s.Equal(expectedID, id)
	})
}
