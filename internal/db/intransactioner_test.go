package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/db"
)

type InTransactionerTestSuite struct {
	dbTestSuite
}

func TestInTransactionerTestSuite(t *testing.T) {
	suite.Run(t, new(InTransactionerTestSuite))
}

func (s *InTransactionerTestSuite) TestInTransaction() {
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")

	s.Run("calls transaction function and returns result from DB", func() {
		d, mock, err := sqlmock.New()
		s.Require().NoError(err)

		mock.ExpectBegin()
		mock.ExpectCommit()

		txnFn := &txnFuncer{req: s.Require()}

		inTxner := db.NewInTransactioner(d)

		err = inTxner.InTransaction(ctx, txnFn.call)

		s.Require().NoError(err)
		s.Require().Equal(1, txnFn.callCount)
	})

	s.Run("returns error from transaction function", func() {
		expectedErr := errors.New("oh no")

		d, mock, err := sqlmock.New()
		s.Require().NoError(err)

		mock.ExpectBegin()
		mock.ExpectRollback()

		txnFn := &txnFuncer{req: s.Require(), err: expectedErr}

		inTxner := db.NewInTransactioner(d)

		err = inTxner.InTransaction(ctx, txnFn.call)

		s.ErrorIs(err, expectedErr)
		s.Require().Equal(1, txnFn.callCount)
	})

	s.Run("returns error on begin transaction failure", func() {
		expectedErr := errors.New("oh no")

		d, mock, err := sqlmock.New()
		s.Require().NoError(err)

		mock.ExpectBegin().WillReturnError(expectedErr)
		mock.ExpectCommit()

		inTxner := db.NewInTransactioner(d)

		err = inTxner.InTransaction(context.Background(), nil)

		s.ErrorIs(err, expectedErr)
	})
}

func (s *InTransactionerTestSuite) wrap(arg interface{}) error {
	if f, ok := arg.(func() error); ok {
		return f()
	}
	s.FailNow("Invalid argument type to Wrap")
	return nil
}

type txnFuncer struct {
	req       *require.Assertions
	callCount int
	err       error
}

func (f *txnFuncer) call(ctx context.Context, tx db.Tx) error {
	f.req.IsType(&sql.Tx{}, tx)
	f.callCount += 1
	return f.err
}
