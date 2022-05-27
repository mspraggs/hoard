package dirscanner

import (
	"io/fs"
	"os"
	"sync"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/util"
)

// DirScannerBuilder implements the builder pattern for DirScanner objects.
type DirScannerBuilder struct {
	fs                fs.FS
	vc                VersionCalculator
	salter            Salter
	bucket            string
	encAlg            models.EncryptionAlgorithm
	numHandlerThreads int
	processor         []Processor
}

// NewBuilder creates a new DirScannerBuilder object with a set of default
// parameters. These parameters are:
//
//   - DirFS filesystem in the current directory (".")
//   - Version calclulator that produces file versions based on their
//     last modification time.
//   - Salter that generates a set of 32 random bytes.
//   - Bucket named "default".
//   - Encryption algorithm set to AES256.
//   - Number of handler threads set to 4.
//   - No handlers.
func NewBuilder() *DirScannerBuilder {
	return &DirScannerBuilder{
		fs:                os.DirFS("."),
		vc:                versionCalculator(calculateVersion),
		salter:            salter(salt),
		bucket:            "default",
		encAlg:            models.EncryptionAlgorithmAES256,
		numHandlerThreads: 4,
	}
}

// WithFS sets the filesystem implementation to be used when the DirScanners is
// created.
func (b *DirScannerBuilder) WithFS(fs fs.FS) *DirScannerBuilder {
	b.fs = fs
	return b
}

// WithVersionCalculator sets the version calculator implementation to be used
// when the DirScanner is created.
func (b *DirScannerBuilder) WithVersionCalculator(vc VersionCalculator) *DirScannerBuilder {
	b.vc = vc
	return b
}

// WithVersionCalculator sets the salter implementation to be used when the
// DirScanner is created.
func (b *DirScannerBuilder) WithSalter(s Salter) *DirScannerBuilder {
	b.salter = s
	return b
}

// WithVersionCalculator sets the bucket to be used when the DirScanner is
// created.
func (b *DirScannerBuilder) WithBucket(bkt string) *DirScannerBuilder {
	b.bucket = bkt
	return b
}

// WithVersionCalculator sets the encryption algorithm to be used when the
// DirScanner is created.
func (b *DirScannerBuilder) WithEncryptionAlgorithm(
	alg models.EncryptionAlgorithm,
) *DirScannerBuilder {

	b.encAlg = alg
	return b
}

// WithVersionCalculator sets the number of handler threads to be used when the
// DirScanner is created.
func (b *DirScannerBuilder) WithNumHandlerThreads(n int) *DirScannerBuilder {
	b.numHandlerThreads = n
	return b
}

// AddProcessor appends a processor instance to the array of processor to be
// used when the DirScanner is created.
func (b *DirScannerBuilder) AddProcessor(h Processor) *DirScannerBuilder {
	b.processor = append(b.processor, h)
	return b
}

// Build constructs a DirScanner instance using the values held by the builder.
func (b *DirScannerBuilder) Build() *DirScanner {
	log := util.MustNewLogger()
	return &DirScanner{
		b.fs,
		b.vc,
		b.salter,
		b.bucket,
		b.encAlg,
		b.numHandlerThreads,
		b.processor,
		nil,
		&sync.WaitGroup{},
		log,
	}
}
