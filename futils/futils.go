package futils

import (
	"runtime"
	"strings"
)

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

var (
	hideFile   func(string) (string, error)
	diskStatus func(string) (DiskStatus, error)
	driveList  func() []string
)

func init() {
	if runtime.GOOS == "windows" {
		hideFile = winHideFile
		diskStatus = winDiskStatus
		driveList = winDriveList
	} else {
		hideFile = unixHideFile
		diskStatus = unixDiskStatus
		driveList = unixDriveList
	}
}

func NormalizePath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	drive := p[0:1]
	rest := p[2:]
	return "/" + strings.ToLower(drive) + strings.Replace(rest, "\\", "/", -1)
}

func UnNormalizePath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	drive := p[1:2]
	rest := p[3:]
	return strings.ToUpper(drive) + ":\\" + strings.Replace(rest, "/", "\\", -1)
}

func HideFile(f string) (string, error) {
	return hideFile(f)
}

func DiskUsage(disk string) (DiskStatus, error) {
	return diskStatus(disk)
}

func DriveList() []string {
	return driveList()
}
