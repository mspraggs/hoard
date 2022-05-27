package app

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/mspraggs/hoard/internal/config"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/dirscanner"
	"github.com/mspraggs/hoard/internal/processor"
	"github.com/mspraggs/hoard/internal/registry"
	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/store/uploader"
	"github.com/mspraggs/hoard/internal/util"
)

// Backup provides the logic to run Hoard's backup functionality.
type Backup struct {
	Command
	ConfigPath       string `required:"true" short:"c" long:"config" description:"The path to the YAML configuration required by hoard"`
	EncryptionSecret string `required:"true" short:"s" long:"secret" env:"HOARD_ENCRYPTION_SECRET" description:"The encryption secret to use when generating encryption keys"`
}

// NewBackup instantiates an instance of the Backup command.
func NewBackup() *Backup {
	return &Backup{}
}

// Execute implements the go-flags Commander interface for the backup command,
// which uses the supplied configuration YAML to back up a set of directories to
// a storage backend.
func (c *Backup) Execute(args []string) error {
	config, err := parseConfig(c.ConfigPath)
	if err != nil {
		return err
	}

	c.configureLogging(&config.Logging)

	client, err := newClient(config)
	if err != nil {
		return err
	}

	d, err := sql.Open("sqlite3", config.Registry.Path)
	if err != nil {
		return err
	}
	d.SetMaxOpenConns(1)

	err = c.uploadFiles(config, d, client)
	if err != nil {
		return err
	}

	return c.storeRegistry(config.Registry, client)
}

func (c *Backup) uploadFiles(config *config.Config, d *sql.DB, client *s3.Client) error {
	defer d.Close()

	inTxner, err := newTransactioner(d)
	if err != nil {
		return err
	}

	for _, dir := range config.Directories {
		err = processDirectory(dir, inTxner, client, config, c.EncryptionSecret)
		if err != nil {
			c.log.Warnw("Unable to process directory", "error", err)
		}
	}

	return nil
}

func (c *Backup) storeRegistry(cfg config.RegConfig, client *s3.Client) error {
	c.log.Infow("Storing registry", "bucket", cfg.Bucket)

	f, err := os.Open(cfg.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	key := filepath.Base(cfg.Path)
	req := &s3.PutObjectInput{
		Bucket: &cfg.Bucket,
		Key:    &key,
		Body:   f,
	}

	_, err = client.PutObject(context.Background(), req)
	if err != nil {
		c.log.Warnw("Failed to store registry", "error", err)
	}
	return err
}

func processDirectory(
	dir config.DirConfig,
	inTxner *db.InTransactioner,
	client *s3.Client,
	config *config.Config,
	secret string,
) error {

	fs := os.DirFS(dir.Path)

	registry := registry.New(
		&util.Clock{}, &inTransactioner{inTxner},
		util.NewRequestIDMaker(),
	)

	store := store.New(
		fs, makeUploaderConstructor(client, config.Uploads.MultiUploadThreshold),
		config.Uploads.ChecksumAlgorithm.ToBusiness(),
		util.NewEncryptionKeyGenerator([]byte(secret)),
		dir.StorageClass.ToBusiness(),
	)

	handler := processor.New(store, registry)

	scanner := dirscanner.NewBuilder().
		WithBucket(dir.Bucket).
		WithFS(fs).
		WithNumHandlerThreads(config.NumThreads).
		WithEncryptionAlgorithm(dir.EncryptionAlgorithm.ToBusiness()).
		WithVersionCalculator(util.NewVersionCalculator(fs)).
		AddProcessor(handler).
		Build()

	return scanner.Scan(context.Background())
}

func makeUploaderConstructor(client *s3.Client, chunksize int64) store.UploaderConstructor {
	return func(file fs.File) (store.Uploader, error) {
		info, err := file.Stat()
		if err != nil {
			return nil, err
		}
		size := info.Size()
		if size >= chunksize {
			return uploader.NewMultiUploader(size, chunksize, client), nil
		}
		return uploader.NewSingleUploader(client), nil
	}
}
