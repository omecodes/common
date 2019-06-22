package futils

import (
	"syscall"
)

func winHideFile(filename string) (string, error) {
	filenameW, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return "", err
	}
	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return "", err
	}
	return filename, nil
}
