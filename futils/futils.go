package futils

import (
	"runtime"
	"strings"
)

var (
	hideFile func(string) (string, error)
)

func init() {
	if runtime.GOOS == "windows" {
		hideFile = winHideFile
	} else {
		hideFile = unixHideFile
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
