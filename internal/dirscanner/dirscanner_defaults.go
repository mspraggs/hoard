package dirscanner

import (
	"crypto/rand"
	"fmt"
	"os"
)

type salter func(string) ([]byte, error)

func (s salter) Salt(path string) ([]byte, error) {
	return s(path)
}

func salt(path string) ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

type versionCalculator func(string) (string, error)

func (c versionCalculator) CalculateVersion(path string) (string, error) {
	return c(path)
}

func calculateVersion(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v_%v", path, info.ModTime()), nil
}
