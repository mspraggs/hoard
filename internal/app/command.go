package app

import (
	"context"
	"database/sql"
	"io/ioutil"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/doug-martin/goqu"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/mspraggs/hoard/internal/config"
	"github.com/mspraggs/hoard/internal/db"
	"github.com/mspraggs/hoard/internal/fileregistry"
	"github.com/mspraggs/hoard/internal/util"
)

// Command provides common based logic for Hoard sub-commands to build upon.
type Command struct {
	getEnv func(string) string
	log    *zap.SugaredLogger
}

func (c *Command) configureLogging(cfg *config.LogConfig) {
	logOptions := []util.LogConfigOption{
		util.WithLogLevel(cfg.Level.ToBusiness()),
	}
	if cfg.FilePath != "" {
		logOptions = append(logOptions, util.WithOutputFilePath(cfg.FilePath))
	}
	util.ConfigureLogging(logOptions...)

	c.log = util.MustNewLogger()
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

func newTransactioner(d *sql.DB) (*db.InTransactioner, error) {
	inTxner := db.NewInTransactioner(goqu.New("sqlite3", d))
	return inTxner, nil
}

type inTransactioner struct {
	inTxner *db.InTransactioner
}

// InTransaction implements the fileregistry InTransactioner interface,
// providing a shim between the interface in that package and the type in the db
// package.
func (t *inTransactioner) InTransaction(
	ctx context.Context,
	fn fileregistry.TxnFunc,
) (interface{}, error) {
	return t.inTxner.InTransaction(
		ctx,
		func(ctx context.Context, s *db.Store) (interface{}, error) {
			return fn(ctx, fileregistry.Store(s))
		},
	)
}
