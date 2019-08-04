package app

import (
	"github.com/shibukawa/configdir"
	"os"
)

type Dir struct {
	path string
}

func (d *Dir) Create() error {
	return os.MkdirAll(d.path, os.ModePerm)
}

func (d *Dir) Path() string {
	return d.path
}

func GetDir(vendor, appName string) *Dir {
	dirs := configdir.New(vendor, appName)
	appData := dirs.QueryFolders(configdir.Global)[0]
	return &Dir{appData.Path}
}
