package util

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"syscall"
	"time"
)

// VersionCalculator provides the logic for calculating the version string of a
// file.
type VersionCalculator struct {
	fs fs.FS
}

// NewVersionCalculator constructs a new version calculator from the provide
// filesystem.
func NewVersionCalculator(fs fs.FS) *VersionCalculator {
	return &VersionCalculator{fs}
}

// CalculateVersion calculates a file version string using the provided file
// path. The version string is derived from the underlying file's ctime.
func (vc *VersionCalculator) CalculateVersion(path string) (string, error) {
	file, err := vc.fs.Open(path)
	if err != nil {
		return "", err
	}

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	ctime, err := ctimeFromFileInfo(info)
	if err != nil {
		return "", err
	}

	versionIdentifier := fmt.Sprintf("%v##%v", path, ctime.UnixNano())
	versionIdentifierHash := md5.Sum([]byte(versionIdentifier))

	return base64.StdEncoding.EncodeToString(versionIdentifierHash[:]), nil
}

func ctimeFromFileInfo(fi fs.FileInfo) (time.Time, error) {
	sys := fi.Sys()
	if sys == nil {
		return time.Time{}, errors.New("unable to get underlying system file info")
	}

	stat, ok := sys.(*syscall.Stat_t)
	if !ok {
		return time.Time{}, fmt.Errorf("system file info type not supported: %T", sys)
	}

	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)), nil
}
