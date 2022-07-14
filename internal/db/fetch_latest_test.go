package db_test

import (
	"context"
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/processor"
)

func (s *RegistryTestSuite) TestFetchLatest() {
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	path := "path/to/file"

	s.Run("fetches latest file row in transaction", func() {
		timestamp := time.Unix(1, 0)
		expectedFile := &processor.File{
			LocalPath: path,
		}
		expectedFileRow := &db.FileRow{
			LocalPath: path,
		}

		clock := fakeClock(func() time.Time { return timestamp })
		s.mockInTransactioner.EXPECT().
			InTransaction(ctx, gomock.Any()).DoAndReturn(fakeInTransaction)
		s.mockLatestFetcher.EXPECT().
			FetchLatest(ctx, gomock.Any(), path).Return(expectedFileRow, nil)

		registry := db.NewRegistry(clock, s.mockInTransactioner, nil, s.mockLatestFetcher, nil)

		latestFile, err := registry.FetchLatest(ctx, path)

		s.Require().NoError(err)
		s.Equal(expectedFile, latestFile)
	})

	s.Run("returns nil when latest fetcher returns nil row", func() {
		timestamp := time.Unix(1, 0)

		clock := fakeClock(func() time.Time { return timestamp })
		s.mockInTransactioner.EXPECT().
			InTransaction(ctx, gomock.Any()).DoAndReturn(fakeInTransaction)
		s.mockLatestFetcher.EXPECT().
			FetchLatest(ctx, gomock.Any(), path).Return(nil, nil)

		registry := db.NewRegistry(clock, s.mockInTransactioner, nil, s.mockLatestFetcher, nil)

		latestFile, err := registry.FetchLatest(ctx, path)

		s.Require().NoError(err)
		s.Nil(latestFile)
	})

	s.Run("handles error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from in transactioner", func() {
			timestamp := time.Unix(1, 0)

			clock := fakeClock(func() time.Time { return timestamp })
			s.mockInTransactioner.EXPECT().
				InTransaction(ctx, gomock.Any()).Return(expectedErr)

			registry := db.NewRegistry(clock, s.mockInTransactioner, nil, nil, nil)

			latestFile, err := registry.FetchLatest(ctx, path)

			s.Nil(latestFile)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from latest fetcher", func() {
			timestamp := time.Unix(1, 0)

			clock := fakeClock(func() time.Time { return timestamp })
			s.mockInTransactioner.EXPECT().
				InTransaction(ctx, gomock.Any()).DoAndReturn(fakeInTransaction)
			s.mockLatestFetcher.EXPECT().
				FetchLatest(ctx, gomock.Any(), path).Return(nil, expectedErr)

			registry := db.NewRegistry(clock, s.mockInTransactioner, nil, s.mockLatestFetcher, nil)

			latestFile, err := registry.FetchLatest(ctx, path)

			s.Nil(latestFile)
			s.ErrorIs(err, expectedErr)
		})
	})
}
