package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

type LatestFetcherTestSuite struct {
	dbTestSuite
}

func TestLatestFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(LatestFetcherTestSuite))
}

func (s *LatestFetcherTestSuite) TestCreate() {
	path := "/some/path"

	rows := []*db.FileRow{
		{
			ID:                 "some-id",
			Key:                "some-key",
			LocalPath:          path,
			Checksum:           42,
			Bucket:             "some-bucket",
			ETag:               "some-etag",
			Version:            "some-version",
			CreatedAtTimestamp: time.Unix(1, 0).UTC(),
		},
		{
			ID:                 "some-other-id",
			Key:                "some-key",
			LocalPath:          path,
			Checksum:           43,
			Bucket:             "some-bucket",
			ETag:               "some-other-etag",
			Version:            "some-version",
			CreatedAtTimestamp: time.Unix(2, 0).UTC(),
		},
	}

	s.Run("returns latest row", func() {
		s.insertFileRow(rows[0])
		s.insertFileRow(rows[1])

		latestFetcher := db.NewGoquLatestFetcher()

		var fetchedRow *db.FileRow
		err := s.inTransaction(func(tx *sql.Tx) error {
			var err error
			fetchedRow, err = latestFetcher.FetchLatest(context.Background(), tx, path)
			if err != nil {
				return err
			}
			return nil
		})

		s.Require().NoError(err)
		s.Equal(rows[1], fetchedRow)
	})

	s.Run("returns nil when no matching rows", func() {
		missingPath := "/path/not/found"
		s.insertFileRow(rows[0])
		s.insertFileRow(rows[1])
		latestFetcher := db.NewGoquLatestFetcher()

		var fetchedRow *db.FileRow
		err := s.inTransaction(func(tx *sql.Tx) error {
			var err error
			fetchedRow, err = latestFetcher.FetchLatest(context.Background(), tx, missingPath)
			if err != nil {
				return err
			}
			return nil
		})

		s.Require().NoError(err)
		s.Nil(fetchedRow)
	})
}
