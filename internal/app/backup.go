package app

import (
	"context"
	"database/sql"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/mspraggs/hoard/internal/config"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/dirscanner"
	"github.com/mspraggs/hoard/internal/processor"
	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/util"
)

// Backup provides the logic to run Hoard's backup functionality.
type Backup struct {
	Command
	ConfigPath string `required:"true" short:"c" long:"config" description:"The path to the YAML configuration required by hoard"`
}

// NewBackup instantiates an instance of the Backup command.
func NewBackup(opts ...CommandOption) *Backup {
	b := &Backup{}

	for _, opt := range opts {
		opt(&b.Command)
	}

	return b
}

// Execute implements the go-flags Commander interface for the backup command,
// which uses the supplied configuration YAML to back up a set of directories to
// a storage backend.
func (b *Backup) Execute(args []string) error {
	config := b.config
	if config == nil {
		var err error
		if config, err = parseConfig(b.ConfigPath); err != nil {
			return err
		}
	}

	b.configureLogging(&config.Logging)

	unlock, err := b.tryLockPID(config.Lockfile)
	if err != nil {
		return err
	}
	defer unlock()

	client, err := newClient(&config.Store)
	if err != nil {
		return err
	}

	d, err := sql.Open("postgres", config.Registry.Location)
	if err != nil {
		return err
	}

	err = b.uploadFiles(config, d, client)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backup) uploadFiles(config *config.Config, d *sql.DB, client *s3.Client) error {
	defer d.Close()

	inTxner := newTransactioner(d)

	for _, dir := range config.Directories {
		if err := processDirectory(
			config.Uploads,
			dir,
			inTxner,
			client,
			config,
		); err != nil {
			b.log.Warnw("Unable to process directory", "error", err)
		}
	}

	return nil
}

func processDirectory(
	uploads config.UploadConfig,
	dir config.DirConfig,
	inTxner db.InTransactioner,
	client *s3.Client,
	config *config.Config,
) error {

	fs := os.DirFS(dir.Path)

	registry := db.NewRegistry(
		&util.Clock{},
		inTxner,
		db.NewCreatorTx(),
		db.NewLatestFetcherTx(),
		rng{},
	)

	store := store.New(
		client,
		fs,
		dir.Bucket,
		store.WithChecksumAlgorithm(uploads.ChecksumAlgorithm.ToInternal()),
		store.WithChunkSize(uploads.MultiUploadThreshold),
		store.WithStorageClass(dir.StorageClass.ToInternal()),
	)

	processor := processor.New(fs, store, registry)

	scanner := dirscanner.New(fs, []dirscanner.Processor{processor}, config.NumThreads)

	return scanner.Scan(context.Background())
}

type rng struct{}

func (g rng) GenerateID() string {
	return uuid.NewString()
}
