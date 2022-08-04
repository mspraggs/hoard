package app

import (
	"context"
	"crypto/tls"
	"database/sql"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nightlyone/lockfile"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/mspraggs/hoard/internal/config"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/util"
)

// CommandOption provides a way to configure a command.
type CommandOption func(*Command)

// Command provides common base logic for Hoard sub-commands to build upon.
type Command struct {
	getEnv   func(string) string
	log      *zap.SugaredLogger
	pidLock  lockfile.Lockfile
	config   *config.Config
	s3Client *s3.Client
}

// WithConfig sets the configuration on the command instance.
func WithConfig(config *config.Config) CommandOption {
	return func(c *Command) {
		c.config = config
	}
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

func newClient(config *config.StoreConfig) (*s3.Client, error) {
	cfgOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(config.Region),
	}

	if creds := config.Credentials; creds != nil {
		cfgOpts = append(
			cfgOpts, awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(creds.ID, creds.Secret, creds.Token),
			),
		)
	}

	if endpoint := config.Endpoint; endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			},
		)
		cfgOpts = append(cfgOpts, awsconfig.WithEndpointResolverWithOptions(resolver))
	}
	if config.DisableTLSChecks {
		client := awshttp.NewBuildableClient()
		client = client.WithTransportOptions(func(t *http.Transport) {
			t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		})
		cfgOpts = append(cfgOpts, awsconfig.WithHTTPClient(client))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, err
	}

	s3Opts := []func(o *s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = config.UsePathStyle
		},
	}
	return s3.NewFromConfig(cfg, s3Opts...), nil
}

func newTransactioner(d *sql.DB) db.InTransactioner {
	return db.NewInTransactioner(d)
}
