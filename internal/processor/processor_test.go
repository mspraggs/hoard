package processor_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/processor/mocks"
)

type contextKey string

type ProcessorTestSuite struct {
	suite.Suite
	controller   *gomock.Controller
	mockRegistry *mocks.MockRegistry
	mockUploader *mocks.MockUploader
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

func (s *ProcessorTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockRegistry = mocks.NewMockRegistry(s.controller)
	s.mockUploader = mocks.NewMockUploader(s.controller)
}
