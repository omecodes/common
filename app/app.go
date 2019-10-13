package app

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/shibukawa/configdir"
	"os"
	"path/filepath"
)

type App struct {
	Vendor   string
	Name     string
	Version  string
	dataDir  string
	confDir  string
	cacheDir string
}

func (a *App) DataDir() (string, error) {
	if a.dataDir == "" {

		hd, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		a.dataDir = filepath.Join(hd, fmt.Sprintf(".%s", a.Vendor), a.Name)
		err = os.MkdirAll(a.dataDir, os.ModePerm)
		if err != nil {
			return a.dataDir, err
		}
	}
	return a.dataDir, nil
}

func (a *App) CacheDir() (string, error) {
	if a.cacheDir == "" {
		dirs := configdir.New(a.Vendor, a.Name)
		appData := dirs.QueryFolders(configdir.Cache)[0]
		a.confDir = appData.Path
		err := os.MkdirAll(a.cacheDir, os.ModePerm)
		if err != nil {
			return a.cacheDir, err
		}
	}
	return a.cacheDir, nil
}

func (a *App) ConfigsDir() (string, error) {
	if a.confDir != "" {
		return a.confDir, nil
	}
	dirs := configdir.New(a.Vendor, a.Name)
	appData := dirs.QueryFolders(configdir.Global)[0]
	a.confDir = appData.Path
	err := os.MkdirAll(a.confDir, os.ModePerm)
	if err != nil {
		return a.confDir, err
	}
	return a.confDir, nil
}
