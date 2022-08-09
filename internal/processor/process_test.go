package processor_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"time"

	"github.com/psanford/memfs"

	"github.com/mspraggs/hoard/internal/processor"
)

type fakeKeyGenerator func() string

func (fn fakeKeyGenerator) GenerateKey() string {
	return fn()
}

type fakeCTimeGetter func() (time.Time, error)

func (fn fakeCTimeGetter) GetCTime(fi fs.File) (time.Time, error) {
	return fn()
}

func (s *ProcessorTestSuite) TestProcess() {
	body := []byte{1, 2, 3}
	key := "key"
	ctime := time.Unix(123, 456).UTC()
	version := "123"
	path := "path/to/file"
	checksum := processor.Checksum(1438416925)
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")

	fs := memfs.New()
	fs.MkdirAll("path/to", os.FileMode(0))
	fs.WriteFile(path, body, os.FileMode(0))

	keyGen := fakeKeyGenerator(func() string { return key })
	ctimeGetter := fakeCTimeGetter(func() (time.Time, error) { return ctime, nil })

	prevFile := &processor.File{
		Key:       key,
		LocalPath: path,
		Checksum:  7,
		CTime:     time.Unix(12, 345).UTC(),
		Version:   "456",
	}
	currentFile := &processor.File{
		Key:       key,
		LocalPath: path,
		Checksum:  checksum,
		CTime:     ctime,
	}
	uploadedFile := &processor.File{
		Key:       key,
		LocalPath: path,
		Checksum:  checksum,
		CTime:     ctime,
		Version:   version,
	}

	s.Run("creates and uploads file", func() {
		s.Run("where file previously uploaded", func() {
			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(prevFile, nil)
			s.mockUploader.EXPECT().Upload(ctx, currentFile).Return(uploadedFile, nil)
			s.mockRegistry.EXPECT().Create(ctx, uploadedFile).Return(uploadedFile, nil)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.Require().NoError(err)
			s.Equal(uploadedFile, file)
		})
		s.Run("where file never uploaded", func() {
			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(nil, nil)
			s.mockUploader.EXPECT().Upload(ctx, currentFile).Return(uploadedFile, nil)
			s.mockRegistry.EXPECT().Create(ctx, uploadedFile).Return(uploadedFile, nil)

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.Require().NoError(err)
			s.Equal(uploadedFile, file)
		})
	})

	s.Run("file not uploaded", func() {
		s.Run("for matching ctime", func() {
			prevFile := &processor.File{
				Key:       key,
				LocalPath: path,
				CTime:     ctime,
				Version:   "456",
			}

			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(prevFile, nil)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.Require().NoError(err)
			s.Nil(file)
		})
		s.Run("for matching checksum", func() {
			prevFile := &processor.File{
				Key:       key,
				LocalPath: path,
				CTime:     time.Unix(12, 345).UTC(),
				Checksum:  checksum,
				Version:   "456",
			}

			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(prevFile, nil)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.Require().NoError(err)
			s.Nil(file)
		})
	})

	s.Run("handles error", func() {
		expectedErr := errors.New("oh no")

		s.Run("when file not found", func() {
			path := "file/not/found"
			processor := processor.New(fs, nil, nil)

			file, err := processor.Process(ctx, path)

			s.ErrorIs(err, os.ErrNotExist)
			s.Nil(file)
		})
		s.Run("from get ctime", func() {
			ctimeGetter := fakeCTimeGetter(func() (time.Time, error) {
				return time.Time{}, expectedErr
			})

			processor := processor.New(
				fs, nil, nil,
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.ErrorIs(err, expectedErr)
			s.Nil(file)
		})
		s.Run("from fetch latest", func() {
			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(nil, expectedErr)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, nil, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.ErrorIs(err, expectedErr)
			s.Nil(file)
		})
		s.Run("from upload", func() {
			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(prevFile, nil)
			s.mockUploader.EXPECT().Upload(ctx, currentFile).Return(nil, expectedErr)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.ErrorIs(err, expectedErr)
			s.Nil(file)
		})
		s.Run("from create", func() {
			s.mockRegistry.EXPECT().FetchLatest(ctx, path).Return(prevFile, nil)
			s.mockUploader.EXPECT().Upload(ctx, currentFile).Return(uploadedFile, nil)
			s.mockRegistry.EXPECT().Create(ctx, uploadedFile).Return(nil, expectedErr)
			keyGen := fakeKeyGenerator(func() string { return "foo" })

			processor := processor.New(
				fs, s.mockUploader, s.mockRegistry,
				processor.WithKeyGenerator(keyGen),
				processor.WithCTimeGetter(ctimeGetter),
			)

			file, err := processor.Process(ctx, path)

			s.ErrorIs(err, expectedErr)
			s.Nil(file)
		})
	})
}
