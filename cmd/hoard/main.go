package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/mspraggs/hoard/internal/app"
)

func main() {
	parser := flags.NewParser(nil, flags.Default)
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
