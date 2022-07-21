package config

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mspraggs/hoard/internal/store"
)

// LogLevel is the YAML configuration representation of the configured log
// level.
type LogLevel string

const (
	LogLevelDebug    LogLevel = "DEBUG"
	LogLevelInfo     LogLevel = "INFO"
	LogLevelWarning  LogLevel = "WARNING"
	LogLevelError    LogLevel = "ERROR"
	LogLevelCritical LogLevel = "DEBUG"
)

// ChecksumAlgorithm is the YAML configuration representation of a configured
// checksum algorithm.
type ChecksumAlgorithm string

const (
	// ChecksumAlgorithmCRC32 is the YAML configuration representation of the
	// CRC32 checksum algorithm.
	ChecksumAlgorithmCRC32 ChecksumAlgorithm = "CRC32"
)

// StorageClass is the YAML configuration representation of a configured storage
// class.
type StorageClass string

const (
	// StorageClassStandard denotes the standard storage backend object class.
	StorageClassStandard StorageClass = "STANDARD"
	// StorageClassArchiveFlexi denotes long-term backend storage with read
	// times ranging from minutes to 12 hours.
	StorageClassArchiveFlexi StorageClass = "ARCHIVE_FLEXI"
	// StorageClassArchiveDeep denotes long-term backend storage with long read
	// times between 12 and 48 hours.
	StorageClassArchiveDeep StorageClass = "ARCHIVE_DEEP"
	// StorageClassArchiveInstant denotes long-term backend storage with instant
	// access reads.
	StorageClassArchiveInstant StorageClass = "ARCHIVE_INSTANT"
)

// Config contains all configuration necessary for the application to run.
type Config struct {
	NumThreads  int          `yaml:"num_threads"`
	Lockfile    string       `yaml:"lock_file"`
	Logging     LogConfig    `yaml:"logging"`
	Registry    RegConfig    `yaml:"registry"`
	Store       StoreConfig  `yaml:"store"`
	Uploads     UploadConfig `yaml:"uploads"`
	Directories []DirConfig  `yaml:"directories"`
}

// LogConfig contains all configuration relating to logs.
type LogConfig struct {
	Level    LogLevel `yaml:"level"`
	FilePath string   `yaml:"file_path"`
}

// RegConfig contains all configuration relating to the file registry.
type RegConfig struct {
	Bucket      string `yaml:"bucket"`
	Path        string `yaml:"path"`
	SaltsRecord string `yaml:"salts_record"`
}

// StoreConfig contains all configuration relating to the file store.
type StoreConfig struct {
	Region string `yaml:"region"`
}

// UploadConfig contains all configuration common to all file uploads.
type UploadConfig struct {
	MultiUploadThreshold int64             `yaml:"multi_upload_threshold"`
	ChecksumAlgorithm    ChecksumAlgorithm `yaml:"checksum_algorithm"`
}

// DirConfig contains all configuration required to configure a directory for
// upload.
type DirConfig struct {
	Bucket       string       `yaml:"bucket"`
	Path         string       `yaml:"path"`
	StorageClass StorageClass `yaml:"storage_class"`
}

// ToInternal converts the YAML representation of a log level to the equivalent
// internal representation.
func (l LogLevel) ToInternal() zapcore.Level {
	switch l {
	case LogLevelDebug:
		return zap.DebugLevel
	case LogLevelInfo:
		return zap.InfoLevel
	case LogLevelWarning:
		return zap.WarnLevel
	case LogLevelError:
		return zap.ErrorLevel
	}
	return zap.InfoLevel
}

// ToInternal converts the YAML represetnation of a checksum algorithm to the
// equivalent internal represenation.
func (a ChecksumAlgorithm) ToInternal() store.ChecksumAlgorithm {
	switch a {
	case ChecksumAlgorithmCRC32:
		return types.ChecksumAlgorithmCrc32
	default:
		return types.ChecksumAlgorithm("")
	}
}

// ToInternal converts the YAML represetnation of a storage class to the
// equivalent internal represenation.
func (c StorageClass) ToInternal() store.StorageClass {
	switch c {
	case StorageClassStandard:
		return types.StorageClassStandard
	case StorageClassArchiveFlexi:
		return types.StorageClassGlacier
	case StorageClassArchiveDeep:
		return types.StorageClassDeepArchive
	case StorageClassArchiveInstant:
		return types.StorageClassGlacierIr
	default:
		return types.StorageClass("")
	}
}
