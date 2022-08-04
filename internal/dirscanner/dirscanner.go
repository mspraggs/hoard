package dirscanner

import (
	"context"
	"io/fs"
	"sync"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/processor"
	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/dirscanner.go -package=mocks -source=$GOFILE

// Processor is the interface required to handle file uploads produced by the
// DirScanner instance.
type Processor interface {
	Process(ctx context.Context, path string) (*processor.File, error)
}

// DirScanner encapsulates the logic for scanning a directory hierarchy and
// generating FileUpload objects from the files within.
type DirScanner struct {
	fs                fs.FS
	numHandlerThreads int
	processors        []Processor
	pathQueue         chan string
	wg                *sync.WaitGroup
	log               *zap.SugaredLogger
}

// New instantiates a new directory scanner instance with the provided options.
func New(fs fs.FS, processors []Processor, numThreads int) *DirScanner {
	return &DirScanner{
		fs:                fs,
		numHandlerThreads: numThreads,
		processors:        processors,
		pathQueue:         make(chan string),
		wg:                &sync.WaitGroup{},
		log:               util.MustNewLogger(),
	}
}

// Scan traverses the filesystem and runs all registered processors on all
// regular files.
func (s *DirScanner) Scan(ctx context.Context) error {
	s.pathQueue = make(chan string)

	for i := 0; i < s.numHandlerThreads; i++ {
		s.wg.Add(1)
		go s.uploadFileUploads(ctx)
	}

	err := fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		s.log.Debugw("Handling path", "path", path)
		if err != nil {
			s.log.Warnw("Skipping file due to error", "error", err, "path", path)
			return nil
		}

		if !d.Type().IsRegular() {
			s.log.Infow("Skipping irregular file type", "path", path, "type", d.Type())
			return nil
		}

		s.pathQueue <- path

		return nil
	})

	close(s.pathQueue)

	s.wg.Wait()

	return err
}

func (s *DirScanner) uploadFileUploads(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case path, ok := <-s.pathQueue:
			if !ok {
				return
			}
			for _, p := range s.processors {
				file, err := p.Process(ctx, path)
				if err != nil {
					s.log.Warnw("Error processing file", "error", err, "path", path)
				} else {
					s.log.Infow("Successfully processed file", "file", file)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
