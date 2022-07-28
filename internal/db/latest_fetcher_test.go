package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

const selectQuery = `
SELECT
	id,
	key,
	local_path,
	checksum,
	bucket,
	etag,
	version,
	created_at_timestamp
FROM files.files
WHERE local_path = \$1
ORDER BY created_at_timestamp DESC
LIMIT 1
`

type LatestFetcherTestSuite struct {
	dbTestSuite
}

func TestLatestFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(LatestFetcherTestSuite))
}

func (s *LatestFetcherTestSuite) TestCreate() {
	path := "/some/path"

	fileRows := []*db.FileRow{
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
		d, mock, err := sqlmock.New()
		s.Require().NoError(err)
		defer d.Close()

		rows := sqlmock.NewRows(insertRows)
		addFileRowsToRows(rows, fileRows[1])

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WithArgs(path).WillReturnRows(rows)
		mock.ExpectCommit()

		latestFetcher := db.NewPostgresLatestFetcher()

		var fetchedRow *db.FileRow
		err = s.inTransaction(d, func(tx *sql.Tx) error {
			var err error
			fetchedRow, err = latestFetcher.FetchLatest(context.Background(), tx, path)
			if err != nil {
				return err
			}
			return nil
		})

		s.Require().NoError(err)
		s.Equal(fileRows[1], fetchedRow)
	})

	s.Run("returns nil when no matching rows", func() {
		missingPath := "/path/not/found"

		d, mock, err := sqlmock.New()
		s.Require().NoError(err)
		defer d.Close()

		rows := sqlmock.NewRows(insertRows)

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WithArgs(missingPath).WillReturnRows(rows)
		mock.ExpectCommit()

		latestFetcher := db.NewPostgresLatestFetcher()

		var fetchedRow *db.FileRow
		err = s.inTransaction(d, func(tx *sql.Tx) error {
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

	s.Run("returns error from scanned row", func() {
		expectedErr := errors.New("fail")

		d, mock, err := sqlmock.New()
		s.Require().NoError(err)
		defer d.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WithArgs(path).WillReturnError(expectedErr)
		mock.ExpectRollback()

		latestFetcher := db.NewPostgresLatestFetcher()

		var fetchedRow *db.FileRow
		err = s.inTransaction(d, func(tx *sql.Tx) error {
			var err error
			fetchedRow, err = latestFetcher.FetchLatest(context.Background(), tx, path)
			if err != nil {
				return err
			}
			return nil
		})

		s.ErrorIs(err, expectedErr)
		s.Nil(fetchedRow)
	})
}
