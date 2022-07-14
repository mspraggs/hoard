package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/db/mocks"
)

type RegistryTestSuite struct {
	suite.Suite
	controller          *gomock.Controller
	mockCreator         *mocks.MockCreator
	mockLatestFetcher   *mocks.MockLatestFetcher
	mockInTransactioner *mocks.MockInTransactioner
}

type MockClock struct {
	now time.Time
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (s *RegistryTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockCreator = mocks.NewMockCreator(s.controller)
	s.mockLatestFetcher = mocks.NewMockLatestFetcher(s.controller)
	s.mockInTransactioner = mocks.NewMockInTransactioner(s.controller)
}

type fakeIDGenerator func() string

func (fn fakeIDGenerator) GenerateID() string {
	return fn()
}

type fakeClock func() time.Time

func (fn fakeClock) Now() time.Time {
	return fn()
}

func fakeInTransaction(ctx context.Context, fn db.TxnFunc) error {
	return fn(ctx, nil)
}
