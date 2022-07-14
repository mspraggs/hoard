package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/mspraggs/hoard/internal/app"
)

var opts struct {
	EncryptionSecret string `required:"true" short:"s" long:"secret" env:"HOARD_ENCRYPTION_SECRET" description:"The encryption secret to use when generating encryption keys"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.AddCommand("backup", "Backup files", "Backup files to AWS S3", app.NewBackup())

	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
