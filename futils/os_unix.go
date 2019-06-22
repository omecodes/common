package futils

import (
	"os"
	"path/filepath"
	"strings"
)

func unixHideFile(filename string) (string, error) {
	if !strings.HasPrefix(filepath.Base(filename), ".") {
		newPath := filepath.Join(filepath.Dir(filename), "."+filepath.Base(filename))
		err := os.Rename(filename, newPath)
		return newPath, err
	}
	return filename, nil
}
