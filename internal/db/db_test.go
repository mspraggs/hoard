package db_test

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/stretchr/testify/suite"
)

type contextKey string

const insertQuery = `
INSERT INTO files (
	id,
	key,
	local_path,
	checksum,
	bucket,
	etag,
	version,
	created_at_timestamp
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
);
`

const selectQuery = `
SELECT * from files;
`

type dbTestSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *dbTestSuite) SetupSuite() {
	s.db = newTestDB()
}

func (s *dbTestSuite) TearDownTest() {
	s.db.Exec("DELETE FROM files")
}

func (s *dbTestSuite) Run(name string, fn func()) {
	s.Suite.Run(name, fn)
	s.TearDownTest()
}

func (s *dbTestSuite) inTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *dbTestSuite) insertFileRow(row *db.FileRow) error {
	_, err := s.db.Exec(
		insertQuery,
		row.ID,
		row.Key,
		row.LocalPath,
		row.Checksum,
		row.Bucket,
		row.ETag,
		row.Version,
		row.CreatedAtTimestamp,
	)
	return err
}

func (s *dbTestSuite) selectFileRows() ([]*db.FileRow, error) {
	rows, err := s.db.Query(selectQuery)
	var fileRows []*db.FileRow
	err = rows.Scan(&fileRows)
	if err != nil {
		return nil, err
	}
	return fileRows, nil
}

func newTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations", "sqlite3", instance,
	)
	if err != nil {
		panic(err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	return db
}
