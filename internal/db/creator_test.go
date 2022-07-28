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

var insertRows = []string{
	"id",
	"key",
	"local_path",
	"checksum",
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
		mock.ExpectQuery("INSERT INTO files.files").WillReturnRows(rows)
		mock.ExpectCommit()

		creator := db.NewPostgresCreator()

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
		mock.ExpectQuery("INSERT INTO files.files").WillReturnRows(rows)
		mock.ExpectCommit()

		creator := db.NewPostgresCreator()

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
