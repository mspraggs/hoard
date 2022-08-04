package app_test

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/app"
	"github.com/mspraggs/hoard/internal/config"
)

const (
	dbName = "hoard"
	dbUser = "hoard"
	dbPass = "password"

	awsID        = "some-account"
	awsSecret    = "some-secret"
	awsToken     = "some-token"
	awsRegion    = "eu-west-2"
	s3BucketName = "test-bucket"

	smallFileSize = 2 * 1024 * 1024  // 2 MiB
	largeFileSize = 50 * 1024 * 1024 // 50 MiB

	numTestFiles = 20

	// Configuration
	numThreads        = 2
	fileSizeThreshold = 10 * 1024 * 1024 // 10 MiB
	checksumAlgorithm = "CRC32"
	storageClass      = config.StorageClassArchiveDeep
)

var lockFile = filepath.Join(os.TempDir(), "backup_test.lock")

type BackupTestSuite struct {
	suite.Suite
}

func TestBackupTestSuite(t *testing.T) {
	rand.Seed(time.Now().Unix())
	suite.Run(t, new(BackupTestSuite))
}

func (s *BackupTestSuite) TestExecute() {
	stop, s3Endpoint := s.setupS3()
	defer stop()

	stop, dbLocation := s.setupDB()
	defer stop()

	directory := s.createTestFiles(numTestFiles, []int{smallFileSize, largeFileSize})
	defer os.RemoveAll(directory)

	cmd := app.NewBackup(app.WithConfig(createHoardConfig(dbLocation, s3Endpoint, directory)))

	err := cmd.Execute([]string{})
	s.Require().NoError(err)

	numDBFiles := s.countDBFiles(dbLocation)
	s.Equal(uint64(numTestFiles), numDBFiles)

	numS3Files := s.countS3Files(s3Endpoint)
	s.Equal(numTestFiles, numS3Files)
}

func (s *BackupTestSuite) setupS3() (func(), string) {
	p := localstack.Preset(localstack.WithServices(localstack.S3))
	ls, err := gnomock.Start(p, gnomock.WithUseLocalImagesFirst(), gnomock.WithDebugMode())
	s.Require().NoError(err)

	s3Endpoint := fmt.Sprintf("https://%s/", ls.Address(localstack.APIPort))
	err = createS3Bucket(s3Endpoint)
	s.Require().NoError(err)

	return func() { gnomock.Stop(ls) }, s3Endpoint
}

func (s *BackupTestSuite) setupDB() (func(), string) {
	p := postgres.Preset(
		postgres.WithDatabase(dbName),
		postgres.WithUser(dbUser, dbPass),
	)
	c, err := gnomock.Start(p, gnomock.WithUseLocalImagesFirst(), gnomock.WithDebugMode())
	s.Require().NoError(err)

	dbLocation := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPass, c.Address("default"), dbName,
	)

	err = applyDBMigrations(dbLocation)
	s.Require().NoError(err)

	return func() { gnomock.Stop(c) }, dbLocation
}

func (s *BackupTestSuite) createTestFiles(numFiles int, fileSizes []int) string {
	directory, err := os.MkdirTemp("", "tmp.*")
	s.Require().NoError(err)

	paths := generateFilePaths(directory, numFiles)

	for _, path := range paths {
		err := os.MkdirAll(filepath.Dir(path), os.ModeDir|0755)
		s.Require().NoError(err)

		fileSize := fileSizes[rand.Intn(len(fileSizes))]
		data := make([]byte, fileSize)
		_, err = rand.Read(data)
		s.Require().NoError(err)
		os.WriteFile(path, data, 0644)
	}

	return directory
}

func (s *BackupTestSuite) countDBFiles(location string) uint64 {
	db, err := sql.Open("postgres", location)
	s.Require().NoError(err)

	result := db.QueryRow("SELECT count(*) FROM files.files;")
	s.Require().NoError(err)

	var count uint64
	err = result.Scan(&count)
	s.Require().NoError(err)

	return count
}

func (s *BackupTestSuite) countS3Files(s3Endpoint string) int {
	client, err := createS3Client(s3Endpoint)
	s.Require().NoError(err)

	output, err := client.ListObjects(
		context.Background(),
		&s3.ListObjectsInput{
			Bucket: aws.String(s3BucketName),
		},
	)
	s.Require().NoError(err)

	return len(output.Contents)
}

func createS3Bucket(s3Endpoint string) error {
	client, err := createS3Client(s3Endpoint)
	if err != nil {
		return err
	}

	_, err = client.CreateBucket(
		context.Background(),
		&s3.CreateBucketInput{
			Bucket: aws.String(s3BucketName),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func createS3Client(endpoint string) (*s3.Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint}, nil
		},
	)

	client := awshttp.NewBuildableClient()
	client = client.WithTransportOptions(func(t *http.Transport) {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	})

	cfgOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(awsID, awsSecret, awsToken),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver),
		awsconfig.WithHTTPClient(client),
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, err
	}

	s3Opts := []func(o *s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = true
		},
	}
	return s3.NewFromConfig(cfg, s3Opts...), nil
}

func applyDBMigrations(location string) error {
	db, err := sql.Open("postgres", location)
	if err != nil {
		return err
	}

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://../../migrations", dbName, driver)
	if err != nil {
		return err
	}

	return m.Up()
}

func generateFilePaths(directory string, numFiles int) []string {
	type queueElem struct {
		path   string
		isFile bool
	}

	queue := []queueElem{{path: directory, isFile: false}}
	uniquePaths := make(map[string]struct{}, numFiles)

	for len(uniquePaths) < numFiles {
		elem := queue[0]
		queue = queue[1:]
		if elem.isFile {
			uniquePaths[elem.path] = struct{}{}
			continue
		}

		for i := 0; i < 4; i++ {
			newElem := queueElem{
				path:   filepath.Join(elem.path, randString(4)),
				isFile: rand.Intn(3) == 0,
			}
			queue = append(queue, newElem)
		}
	}

	paths := make([]string, 0, len(uniquePaths))
	for p := range uniquePaths {
		paths = append(paths, p)
		if len(paths) == numFiles {
			break
		}
	}

	return paths
}

func createHoardConfig(dbLocation, s3Endpoint, directory string) *config.Config {
	return &config.Config{
		NumThreads: numThreads,
		Lockfile:   lockFile,
		Logging: config.LogConfig{
			Level: "DEBUG",
		},
		Registry: config.RegConfig{
			Location: dbLocation,
		},
		Store: config.StoreConfig{
			Region:           awsRegion,
			Endpoint:         s3Endpoint,
			UsePathStyle:     true,
			DisableTLSChecks: true,
			Credentials: &config.StoreCredentials{
				ID:     awsID,
				Secret: awsSecret,
				Token:  awsToken,
			},
		},
		Uploads: config.UploadConfig{
			MultiUploadThreshold: fileSizeThreshold,
			ChecksumAlgorithm:    checksumAlgorithm,
		},
		Directories: []config.DirConfig{
			{
				Bucket:       s3BucketName,
				Path:         directory,
				StorageClass: storageClass,
			},
		},
	}
}

func randString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
