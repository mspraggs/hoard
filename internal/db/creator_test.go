package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

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
		creator := db.NewGoquCreator()

		var insertedRow *db.FileRow
		err := s.inTransaction(func(tx *sql.Tx) error {
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
		err := s.insertFileRow(row)
		s.Require().NoError(err)

		creator := db.NewGoquCreator()

		var insertedRow *db.FileRow
		err = s.inTransaction(func(tx *sql.Tx) error {
			var err error
			insertedRow, err = creator.Create(context.Background(), tx, row)
			if err != nil {
				return err
			}
			return nil
		})

		s.ErrorContains(err, "constraint failed")
		s.Nil(insertedRow)
	})
}
