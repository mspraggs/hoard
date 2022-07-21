package app

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

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
func NewBackup() *Backup {
	return &Backup{}
}

// Execute implements the go-flags Commander interface for the backup command,
// which uses the supplied configuration YAML to back up a set of directories to
// a storage backend.
func (b *Backup) Execute(args []string) error {
	config, err := parseConfig(b.ConfigPath)
	if err != nil {
		return err
	}

	b.configureLogging(&config.Logging)

	unlock, err := b.tryLockPID(config.Lockfile)
	if err != nil {
		return err
	}
	defer unlock()

	client, err := newClient(config)
	if err != nil {
		return err
	}

	d, err := sql.Open("sqlite3", config.Registry.Path)
	if err != nil {
		return err
	}
	d.SetMaxOpenConns(1)

	err = b.uploadFiles(config, d, client)
	if err != nil {
		return err
	}

	return b.storeRegistry(config.Registry, client)
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

func (b *Backup) storeRegistry(cfg config.RegConfig, client *s3.Client) error {
	b.log.Infow("Storing registry", "bucket", cfg.Bucket)

	if err := b.uploadFile(cfg.Path, cfg.Bucket, client); err != nil {
		b.log.Warnw("Failed to store salt record", "error", err)
		return err
	}

	if err := b.uploadFile(
		cfg.Path,
		cfg.Bucket,
		client,
	); err != nil {
		b.log.Warnw("Failed to store registry", "error", err)
		return err
	}
	return nil
}

func (b *Backup) uploadFile(
	path,
	bucket string,
	client *s3.Client,
) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	key := filepath.Base(path)
	req := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	}

	_, err = client.PutObject(context.Background(), req)
	if err != nil {
		b.log.Warnw("Failed to store registry", "error", err)
		return err
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
		db.NewGoquCreator(),
		db.NewGoquLatestFetcher(),
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

	handler := processor.New(fs, rng{}, store, registry)

	scanner := dirscanner.New(fs, []dirscanner.Processor{handler}, config.NumThreads)

	return scanner.Scan(context.Background())
}

type rng struct{}

func (g rng) GenerateID() string {
	return uuid.NewString()
}

func (g rng) GenerateKey() string {
	return g.GenerateID()
}
