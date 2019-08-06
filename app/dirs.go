package app

import (
	"github.com/shibukawa/configdir"
	"os"
)

type ConfDir struct {
	path string
}

func (d *ConfDir) Create() error {
	return os.MkdirAll(d.path, os.ModePerm)
}

func (d *ConfDir) Path() string {
	return d.path
}

func GetDir(vendor, appName string) *ConfDir {
	dirs := configdir.New(vendor, appName)
	appData := dirs.QueryFolders(configdir.Global)[0]
	return &ConfDir{appData.Path}
}
