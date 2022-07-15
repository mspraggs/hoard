package app

import (
	"context"
	"database/sql"
	"io/ioutil"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nightlyone/lockfile"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/mspraggs/hoard/internal/config"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/util"
)

// Command provides common based logic for Hoard sub-commands to build upon.
type Command struct {
	getEnv  func(string) string
	log     *zap.SugaredLogger
	pidLock lockfile.Lockfile
}

func (c *Command) configureLogging(cfg *config.LogConfig) {
	logOptions := []util.LogConfigOption{
		util.WithLogLevel(cfg.Level.ToInternal()),
	}
	if cfg.FilePath != "" {
		logOptions = append(logOptions, util.WithOutputFilePath(cfg.FilePath))
	}
	util.ConfigureLogging(logOptions...)

	c.log = util.MustNewLogger()
}

func (c *Command) tryLockPID(lockfilePath string) (func() error, error) {
	lock, err := lockfile.New(lockfilePath)
	if err != nil {
		return nil, err
	}
	c.pidLock = lock
	if err := c.pidLock.TryLock(); err != nil {
		c.log.Errorw("Failed to acquire PID lock", "lockfile", lockfilePath, "error", err)
		return nil, err
	}
	c.log.Infow("Acquired PID lock", "lockfile", lockfilePath)
	return func() error { return c.pidLock.Unlock() }, nil
}

func parseConfig(path string) (*config.Config, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &config.Config{}
	err = yaml.Unmarshal(raw, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func newClient(config *config.Config) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(config.Store.Region),
	)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

func newTransactioner(d *sql.DB) db.InTransactioner {
	return db.NewInTransactioner(d)
}
