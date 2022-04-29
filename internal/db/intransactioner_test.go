package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/doug-martin/goqu"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/models"
)

type InTransactionerTestSuite struct {
	suite.Suite
	db *goqu.Database
}

func TestInTransactionerTestSuite(t *testing.T) {
	suite.Run(t, new(InTransactionerTestSuite))
}

func (s *InTransactionerTestSuite) SetupSuite() {
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

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	s.db = goqu.New("sqlite3", db)
}

func (s *InTransactionerTestSuite) TestInTransaction() {
	s.Run("calls transaction function and returns result from DB", func() {
		fileUpload := &models.FileUpload{ID: "foo"}

		inTxner := db.NewInTransactioner(s.db)

		opaqueResult, err := inTxner.InTransaction(
			context.Background(),
			func(ctx context.Context, s *db.Store) (interface{}, error) {
				return fileUpload, nil
			},
		)

		s.Require().NoError(err)
		s.Equal(fileUpload, opaqueResult)
	})

	s.Run("returns error from transaction function", func() {
		expectedErr := errors.New("oh no")
		inTxner := db.NewInTransactioner(s.db)

		opaqueResult, err := inTxner.InTransaction(
			context.Background(),
			func(ctx context.Context, s *db.Store) (interface{}, error) {
				return nil, expectedErr
			},
		)

		s.Require().Nil(opaqueResult)
		s.ErrorIs(expectedErr, err)
	})

	s.Run("returns error on begin transaction failure", func() {
		inTxner := db.NewInTransactioner(s.db)

		s.db.Db.Close()
		defer s.SetupSuite()

		opaqueResult, err := inTxner.InTransaction(
			context.Background(),
			func(ctx context.Context, s *db.Store) (interface{}, error) {
				return nil, nil
			},
		)

		s.Require().Nil(opaqueResult)
		s.ErrorContains(err, "is closed")
	})
}
