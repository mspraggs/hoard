package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

const insertQuery = `
INSERT INTO files.files \(
	id,
	key,
	local_path,
	checksum,
	change_time,
	bucket,
	etag,
	version,
	created_at_timestamp
\) VALUES \(
	\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9
\)
RETURNING id, key, local_path, checksum, change_time, bucket, etag, version, created_at_timestamp
`

var insertRows = []string{
	"id",
	"key",
	"local_path",
	"checksum",
	"change_time",
	"bucket",
	"etag",
	"version",
	"created_at_timestamp",
}

type CreatorTestSuite struct {
	dbTestSuite
}

func TestCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(CreatorTestSuite))
}

func (s *CreatorTestSuite) TestCreate() {
	row := &db.FileRow{
		ID:                 "some-id",
		Key:                "some-key",
		LocalPath:          "/some/path",
		Checksum:           42,
		CTime:              time.Unix(123, 456).UTC(),
		Bucket:             "some-bucket",
		ETag:               "some-etag",
		Version:            "some-version",
		CreatedAtTimestamp: time.Unix(1, 0).UTC(),
	}

	s.Run("inserts provided row", func() {
		d, mock, err := sqlmock.New()
		s.Require().NoError(err)
		defer d.Close()

		rows := sqlmock.NewRows(insertRows)
		addFileRowsToRows(rows, row)

		mock.ExpectBegin()
		mock.ExpectQuery(insertQuery).WillReturnRows(rows)
		mock.ExpectCommit()

		creator := db.NewCreatorTx()

		var insertedRow *db.FileRow
		err = s.inTransaction(d, func(tx *sql.Tx) error {
			var err error
			insertedRow, err = creator.Create(context.Background(), tx, row)
			if err != nil {
				return err
			}
			return nil
		})

		s.Require().NoError(err)
		s.Equal(row, insertedRow)
	})

	s.Run("handles error from transaction", func() {
		d, mock, err := sqlmock.New()
		s.Require().NoError(err)
		defer d.Close()

		rows := sqlmock.NewRows(insertRows)

		mock.ExpectBegin()
		mock.ExpectQuery(insertQuery).WillReturnRows(rows)
		mock.ExpectCommit()

		creator := db.NewCreatorTx()

		var insertedRow *db.FileRow
		err = s.inTransaction(d, func(tx *sql.Tx) error {
			var err error
			insertedRow, err = creator.Create(context.Background(), tx, row)
			if err != nil {
				return err
			}
			return nil
		})

		s.ErrorIs(err, sql.ErrNoRows)
		s.Nil(insertedRow)
	})
}
