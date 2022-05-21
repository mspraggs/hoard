package config

import (
	"time"

	"github.com/mspraggs/hoard/internal/models"
)

// ChecksumAlgorithm is the YAML configuration representation of a configured
// checksum algorithm.
type ChecksumAlgorithm string

const (
	// ChecksumAlgorithmSHA256 is the YAML configuration representation of the
	// SHA256 checksum algorithm.
	ChecksumAlgorithmSHA256 ChecksumAlgorithm = "SHA256"
)

// EncryptionAlgorithm is the YAML configuration representation of a configured
// encryption algorithm.
type EncryptionAlgorithm string

const (
	// EncryptionAlgorithmAES256 is the YAML configuration representation of the
	// AES256 encryption algorithm.
	EncryptionAlgorithmAES256 EncryptionAlgorithm = "AES256"
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
	Registry    RegConfig    `yaml:"registry"`
	Uploads     UploadConfig `yaml:"uploads"`
	Directories []DirConfig  `yaml:"directories"`
}

// RegConfig contains all configuration relating to the file registry.
type RegConfig struct {
	Path string `yaml:"path"`
}

// UploadConfig contains all configuration common to all file uploads.
type UploadConfig struct {
	MultiUploadThreshold int64             `yaml:"multi_upload_threshold"`
	ChecksumAlgorithm    ChecksumAlgorithm `yaml:"checksum_algorithm"`
}

// DirConfig contains all configuration required to configure a directory for
// upload.
type DirConfig struct {
	Bucket              string              `yaml:"bucket"`
	Path                string              `yaml:"path"`
	EncryptionAlgorithm EncryptionAlgorithm `yaml:"encryption_algorithm"`
	StorageClass        StorageClass        `yaml:"storage_class"`
	RetentionPeriod     time.Duration       `yaml:"retention_period"`
}

// ToBusiness converts the YAML represetnation of a checksum algorithm to the
// equivalent buisness model representation.
func (a ChecksumAlgorithm) ToBusiness() models.ChecksumAlgorithm {
	switch a {
	case ChecksumAlgorithmSHA256:
		return models.ChecksumAlgorithmSHA256
	default:
		return models.ChecksumAlgorithm(0)
	}
}

// ToBusiness converts the YAML represetnation of a encryption algorithm to the
// equivalent buisness model representation.
func (a EncryptionAlgorithm) ToBusiness() models.EncryptionAlgorithm {
	switch a {
	case EncryptionAlgorithmAES256:
		return models.EncryptionAlgorithmAES256
	default:
		return models.EncryptionAlgorithm(0)
	}
}

// ToBusiness converts the YAML represetnation of a encryption algorithm to the
// equivalent buisness model representation.
func (c StorageClass) ToBusiness() models.StorageClass {
	switch c {
	case StorageClassStandard:
		return models.StorageClassStandard
	case StorageClassArchiveFlexi:
		return models.StorageClassArchiveFlexi
	case StorageClassArchiveDeep:
		return models.StorageClassArchiveDeep
	case StorageClassArchiveInstant:
		return models.StorageClassArchiveInstant
	default:
		return models.StorageClass(0)
	}
}
