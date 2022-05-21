package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/doug-martin/goqu"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/db"
	dbmodels "github.com/mspraggs/hoard/internal/db/models"
	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/models"
)

type StoreTesteSuite struct {
	suite.Suite
	db *goqu.Database
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTesteSuite))
}

func (s *StoreTesteSuite) SetupSuite() {
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

func (s *StoreTesteSuite) TearDownTest() {
	s.db.Exec("DELETE FROM file_uploads")
	s.db.Exec("DELETE FROM file_uploads_history")
}

func (s *StoreTesteSuite) Run(name string, fn func()) {
	s.Suite.Run(name, fn)
	s.TearDownTest()
}

func (s *StoreTesteSuite) TestGetFileByChangeRequestID() {
	requestID := "request-id"
	ID := "foo"
	dbChangeType := dbmodels.ChangeTypeCreate
	insertedFileUploadHistoryRow := &dbmodels.FileUploadHistoryRow{
		RequestID:  requestID,
		ID:         ID,
		ChangeType: dbChangeType,
	}

	changeType := models.ChangeTypeCreate
	fileUpload := &models.FileUpload{
		ID: ID,
	}

	s.Require().NoError(s.insertFileUploadHistoryRow(insertedFileUploadHistoryRow))

	s.Run("returns existing file upload matching request ID", func() {
		expectedChangeType := changeType
		expectedFileUpload := fileUpload

		tx, err := s.db.Begin()
		s.Require().NoError(err)

		var fileUpload *models.FileUpload
		var changeType models.ChangeType
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			fileUpload, changeType, err = store.GetFileUploadByChangeRequestID(
				context.Background(), requestID,
			)

			return err
		})

		s.Require().NoError(err)
		s.Equal(expectedFileUpload, fileUpload)
		s.Equal(expectedChangeType, changeType)
	})

	s.Run("forwards error from DB", func() {
		expectedChangeType := models.ChangeType(0)

		tx, err := s.db.Begin()
		tx.Commit()
		s.Require().NoError(err)

		var fileUpload *models.FileUpload
		var changeType models.ChangeType
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			fileUpload, changeType, err = store.GetFileUploadByChangeRequestID(
				context.Background(), requestID,
			)

			return err
		})

		s.ErrorContains(err, "already been committed")
		s.Nil(fileUpload)
		s.Equal(expectedChangeType, changeType)
	})

	s.Run("returns not found error for missing resource", func() {
		expectedErr := pkgerrors.ErrNotFound
		expectedChangeType := models.ChangeType(0)

		tx, err := s.db.Begin()
		s.Require().NoError(err)

		var fileUpload *models.FileUpload
		var changeType models.ChangeType
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			fileUpload, changeType, err = store.GetFileUploadByChangeRequestID(
				context.Background(), "doesnt-exist",
			)

			return err
		})

		s.Equal(expectedErr, err)
		s.Nil(fileUpload)
		s.Equal(expectedChangeType, changeType)
	})
}

func (s *StoreTesteSuite) TestInsertFileUpload() {
	requestID := "request-id"
	ID := "foo"
	fileUpload := &models.FileUpload{ID: ID}

	s.Run("inserts into file uploads and file uploads history tables", func() {
		s.Run("using existing ID", func() {
			expectedFileUpload := fileUpload
			expectedFileUploadRows := []*dbmodels.FileUploadRow{
				{ID: ID},
			}
			expectedFileUploadHistoryRows := []*dbmodels.FileUploadHistoryRow{
				{
					RequestID:  requestID,
					ID:         ID,
					ChangeType: dbmodels.ChangeTypeCreate,
				},
			}

			tx, err := s.db.Begin()
			s.Require().NoError(err)

			var insertedFileUpload *models.FileUpload
			err = tx.Wrap(func() error {
				store := db.NewStore(tx)
				insertedFileUpload, err = store.InsertFileUpload(
					context.Background(), requestID, fileUpload,
				)

				return err
			})

			s.Require().NoError(err)
			s.Equal(expectedFileUpload, insertedFileUpload)

			fileUploadRows, err := s.selectFileUploadRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadRows, fileUploadRows)

			fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadHistoryRows, fileUploadHistoryRows)
		})
		s.Run("and populates ID when empty", func() {
			fileUpload := &models.FileUpload{}
			expectedFileUpload := &models.FileUpload{}

			tx, err := s.db.Begin()
			s.Require().NoError(err)

			var insertedFileUpload *models.FileUpload
			err = tx.Wrap(func() error {
				store := db.NewStore(tx)
				insertedFileUpload, err = store.InsertFileUpload(
					context.Background(), requestID, fileUpload,
				)

				return err
			})

			s.Require().NoError(err)
			s.isUUID(insertedFileUpload.ID)
			expectedFileUpload.ID = insertedFileUpload.ID
			s.Equal(expectedFileUpload, insertedFileUpload)

			expectedFileUploadRows := []*dbmodels.FileUploadRow{
				{ID: insertedFileUpload.ID},
			}
			expectedFileUploadHistoryRows := []*dbmodels.FileUploadHistoryRow{
				{
					RequestID:  requestID,
					ChangeType: dbmodels.ChangeTypeCreate,
					ID:         insertedFileUpload.ID,
				},
			}

			fileUploadRows, err := s.selectFileUploadRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadRows, fileUploadRows)

			fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadHistoryRows, fileUploadHistoryRows)
		})
	})

	s.Run("forwards error from DB", func() {
		tx, err := s.db.Begin()
		s.Require().NoError(err)
		tx.Commit()

		var insertedFileUpload *models.FileUpload
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			insertedFileUpload, err = store.InsertFileUpload(
				context.Background(), requestID, fileUpload,
			)

			return err
		})

		s.ErrorContains(err, "already been committed")
		s.Nil(insertedFileUpload)

		fileUploadRows, err := s.selectFileUploadRows()
		s.Require().NoError(err)
		s.Empty(fileUploadRows)

		fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
		s.Require().NoError(err)
		s.Empty(fileUploadHistoryRows)
	})

	s.Run("returns error", func() {
		s.Run("when inserting duplicate history row", func() {
			existingFileUploadHistoryRow := &dbmodels.FileUploadHistoryRow{
				RequestID:  requestID,
				ID:         ID,
				ChangeType: dbmodels.ChangeTypeCreate,
			}
			expectedFileUploadHistoryRows := []*dbmodels.FileUploadHistoryRow{
				existingFileUploadHistoryRow,
			}

			err := s.insertFileUploadHistoryRow(existingFileUploadHistoryRow)
			s.Require().NoError(err)

			tx, err := s.db.Begin()
			s.Require().NoError(err)

			var insertedFileUpload *models.FileUpload
			err = tx.Wrap(func() error {
				store := db.NewStore(tx)
				insertedFileUpload, err = store.InsertFileUpload(
					context.Background(), requestID, fileUpload,
				)

				return err
			})

			s.Require().ErrorContains(err, "UNIQUE constraint failed")
			s.Nil(insertedFileUpload)

			fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadHistoryRows, fileUploadHistoryRows)

			fileUploadRows, err := s.selectFileUploadRows()
			s.Require().NoError(err)
			s.Empty(fileUploadRows)
		})
		s.Run("when inserting duplicate row", func() {
			existingFileUploadRow := &dbmodels.FileUploadRow{
				ID: ID,
			}
			expectedFileUploadRows := []*dbmodels.FileUploadRow{
				existingFileUploadRow,
			}

			err := s.insertFileUploadRow(existingFileUploadRow)
			s.Require().NoError(err)

			tx, err := s.db.Begin()
			s.Require().NoError(err)

			var insertedFileUpload *models.FileUpload
			err = tx.Wrap(func() error {
				store := db.NewStore(tx)
				insertedFileUpload, err = store.InsertFileUpload(
					context.Background(), requestID, fileUpload,
				)

				return err
			})

			s.Require().ErrorContains(err, "UNIQUE constraint failed")
			s.Nil(insertedFileUpload)

			fileUploadRows, err := s.selectFileUploadRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadRows, fileUploadRows)

			fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
			s.Require().NoError(err)
			s.Empty(fileUploadHistoryRows)
		})
	})
}

func (s *StoreTesteSuite) TestUpdateFileUpload() {
	requestID := "request-id"
	ID := "foo"
	uploadedAtTimestamp := time.Unix(1, 0).UTC()
	initialFileUploadRow := &dbmodels.FileUploadRow{
		ID: ID,
	}
	initialFileUploadHistoryRow := &dbmodels.FileUploadHistoryRow{
		RequestID:  "initial",
		ChangeType: dbmodels.ChangeTypeCreate,
	}
	newFileUpload := &models.FileUpload{
		ID:                  ID,
		UploadedAtTimestamp: uploadedAtTimestamp,
	}

	s.Run("updates file uploads table and inserts into file uploads history table", func() {
		s.insertFileUploadRow(initialFileUploadRow)
		s.insertFileUploadHistoryRow(initialFileUploadHistoryRow)

		expectedFileUpload := newFileUpload
		expectedFileUploadRows := []*dbmodels.FileUploadRow{
			{
				ID:                  ID,
				UploadedAtTimestamp: uploadedAtTimestamp,
			},
		}
		expectedFileUploadHistoryRows := []*dbmodels.FileUploadHistoryRow{
			initialFileUploadHistoryRow,
			{
				RequestID:           requestID,
				ID:                  ID,
				ChangeType:          dbmodels.ChangeTypeUpdate,
				UploadedAtTimestamp: uploadedAtTimestamp,
			},
		}

		tx, err := s.db.Begin()
		s.Require().NoError(err)

		var updatedFileUpload *models.FileUpload
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			updatedFileUpload, err = store.UpdateFileUpload(
				context.Background(), requestID, newFileUpload,
			)

			return err
		})

		s.Require().NoError(err)
		s.Equal(expectedFileUpload, updatedFileUpload)

		fileUploadRows, err := s.selectFileUploadRows()
		s.Require().NoError(err)
		s.ElementsMatch(expectedFileUploadRows, fileUploadRows)

		fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
		s.Require().NoError(err)
		s.ElementsMatch(expectedFileUploadHistoryRows, fileUploadHistoryRows)
	})

	s.Run("forwards error from DB", func() {
		tx, err := s.db.Begin()
		s.Require().NoError(err)
		tx.Commit()

		var insertedFileUpload *models.FileUpload
		err = tx.Wrap(func() error {
			store := db.NewStore(tx)
			insertedFileUpload, err = store.UpdateFileUpload(
				context.Background(), requestID, newFileUpload,
			)

			return err
		})

		s.ErrorContains(err, "already been committed")
		s.Nil(insertedFileUpload)

		fileUploadRows, err := s.selectFileUploadRows()
		s.Require().NoError(err)
		s.Empty(fileUploadRows)

		fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
		s.Require().NoError(err)
		s.Empty(fileUploadHistoryRows)
	})

	s.Run("returns error", func() {
		s.Run("when inserting duplicate history row", func() {
			s.insertFileUploadRow(initialFileUploadRow)
			s.insertFileUploadHistoryRow(initialFileUploadHistoryRow)

			existingFileUploadHistoryRow := &dbmodels.FileUploadHistoryRow{
				RequestID:  requestID,
				ID:         ID,
				ChangeType: dbmodels.ChangeTypeCreate,
			}
			expectedFileUploadHistoryRows := []*dbmodels.FileUploadHistoryRow{
				initialFileUploadHistoryRow,
				existingFileUploadHistoryRow,
			}
			expectedFileUploadRows := []*dbmodels.FileUploadRow{
				initialFileUploadRow,
			}

			err := s.insertFileUploadHistoryRow(existingFileUploadHistoryRow)
			s.Require().NoError(err)

			tx, err := s.db.Begin()
			s.Require().NoError(err)

			var insertedFileUpload *models.FileUpload
			err = tx.Wrap(func() error {
				store := db.NewStore(tx)
				insertedFileUpload, err = store.UpdateFileUpload(
					context.Background(), requestID, newFileUpload,
				)

				return err
			})

			s.Require().ErrorContains(err, "UNIQUE constraint failed")
			s.Nil(insertedFileUpload)

			fileUploadHistoryRows, err := s.selectFileUploadHistoryRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadHistoryRows, fileUploadHistoryRows)

			fileUploadRows, err := s.selectFileUploadRows()
			s.Require().NoError(err)
			s.ElementsMatch(expectedFileUploadRows, fileUploadRows)
		})
	})
}

func (s *StoreTesteSuite) insertFileUploadRow(row *dbmodels.FileUploadRow) error {
	_, err := s.db.From("file_uploads").Insert(row).Exec()
	return err
}

func (s *StoreTesteSuite) insertFileUploadHistoryRow(row *dbmodels.FileUploadHistoryRow) error {
	_, err := s.db.From("file_uploads_history").Insert(row).Exec()
	return err
}

func (s *StoreTesteSuite) selectFileUploadRows() ([]*dbmodels.FileUploadRow, error) {
	rows := []*dbmodels.FileUploadRow{}
	err := s.db.From("file_uploads").Select(goqu.Star()).ScanStructs(&rows)
	return rows, err
}

func (s *StoreTesteSuite) selectFileUploadHistoryRows() ([]*dbmodels.FileUploadHistoryRow, error) {
	rows := []*dbmodels.FileUploadHistoryRow{}
	err := s.db.From("file_uploads_history").Select(goqu.Star()).ScanStructs(&rows)
	return rows, err
}

func (s *StoreTesteSuite) isUUID(str string) {
	if _, err := uuid.Parse(str); err != nil {
		s.FailNow(fmt.Sprintf("%q is not a valid UUID", str))
	}
}
