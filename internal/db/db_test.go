package db_test

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

type contextKey string

type dbTestSuite struct {
	suite.Suite
}

func (s *dbTestSuite) inTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func addFileRowsToRows(rows *sqlmock.Rows, fileRows ...*db.FileRow) {
	for _, row := range fileRows {
		rows.AddRow(
			row.ID,
			row.Key,
			row.LocalPath,
			row.Checksum,
			row.Bucket,
			row.ETag,
			row.Version,
			row.CreatedAtTimestamp,
		)
	}
}
