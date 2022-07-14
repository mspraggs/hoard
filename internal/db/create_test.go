package db_test

import (
	"context"
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/processor"
)

func (s *RegistryTestSuite) TestCreate() {
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	id := "some-id"
	key := "some-key"

	idGen := fakeIDGenerator(func() string { return id })

	s.Run("creates file row in transaction", func() {
		timestamp := time.Unix(1, 0)
		inputFile := &processor.File{
			Key: key,
		}
		expectedFile := &processor.File{
			Key: key,
		}
		expectedFileRow := &db.FileRow{
			ID:                 id,
			Key:                key,
			CreatedAtTimestamp: timestamp,
		}

		clock := fakeClock(func() time.Time { return timestamp })
		s.mockInTransactioner.EXPECT().
			InTransaction(ctx, gomock.Any()).DoAndReturn(fakeInTransaction)
		s.mockCreator.EXPECT().
			Create(ctx, gomock.Any(), expectedFileRow).Return(expectedFileRow, nil)

		registry := db.NewRegistry(clock, s.mockInTransactioner, s.mockCreator, nil, idGen)

		createdFile, err := registry.Create(ctx, inputFile)

		s.Require().NoError(err)
		s.Equal(expectedFile, createdFile)
	})

	s.Run("handles error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from in transactioner", func() {
			timestamp := time.Unix(1, 0)
			inputFile := &processor.File{
				Key: key,
			}

			clock := fakeClock(func() time.Time { return timestamp })
			s.mockInTransactioner.EXPECT().
				InTransaction(ctx, gomock.Any()).Return(expectedErr)

			registry := db.NewRegistry(clock, s.mockInTransactioner, nil, nil, idGen)

			createdFile, err := registry.Create(ctx, inputFile)

			s.Nil(createdFile)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from creator", func() {
			timestamp := time.Unix(1, 0)
			inputFile := &processor.File{
				Key: key,
			}
			expectedFileRow := &db.FileRow{
				ID:                 id,
				Key:                key,
				CreatedAtTimestamp: timestamp,
			}

			clock := fakeClock(func() time.Time { return timestamp })
			s.mockInTransactioner.EXPECT().
				InTransaction(ctx, gomock.Any()).DoAndReturn(fakeInTransaction)
			s.mockCreator.EXPECT().
				Create(ctx, gomock.Any(), expectedFileRow).Return(nil, expectedErr)

			registry := db.NewRegistry(clock, s.mockInTransactioner, s.mockCreator, nil, idGen)

			createdFile, err := registry.Create(ctx, inputFile)

			s.Nil(createdFile)
			s.ErrorIs(err, expectedErr)
		})
	})
}
