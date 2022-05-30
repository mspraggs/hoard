package processor_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/processor/mocks"
)

type ProcessorTestSuite struct {
	suite.Suite
	controller   *gomock.Controller
	now          time.Time
	mockRegistry *mocks.MockRegistry
	mockStore    *mocks.MockStore
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

func (s *ProcessorTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.now = time.Time{}
	s.mockRegistry = mocks.NewMockRegistry(s.controller)
	s.mockStore = mocks.NewMockStore(s.controller)
}
