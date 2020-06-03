package app

import (
	"github.com/zoenion/common/jcon"
)

type options struct {
	startCMDFunc         func()
	afterConfigure       func(cfg jcon.Map) error
	version              string
	withResources        bool
	configItems          []configItem
	customAppDataDirPath string
	instanceName         string
}

type Option func(*options)

func WithVersion(version string) Option {
	return func(opts *options) {
		opts.version = version
	}
}

func WithRunCommandFunc(f func()) Option {
	return func(opts *options) {
		opts.startCMDFunc = f
	}
}

func WithAfterConfigure(f func(cfg jcon.Map) error) Option {
	return func(opts *options) {
		opts.afterConfigure = f
	}
}

func WithCustomAppData(dirname string) Option {
	return func(opts *options) {
		opts.customAppDataDirPath = dirname
	}
}

func WithResourcesEnabled(enabled bool) Option {
	return func(i *options) {
		i.withResources = enabled
	}
}

func WithConfig(description string, confType ConfigType) Option {
	return func(o *options) {
		o.configItems = append(o.configItems, configItem{description: description, configType: confType})
	}
}

func WithDirConfigs(description string, names ...string) Option {
	return func(o *options) {
		o.configItems = append(o.configItems, configItem{description: description, configType: ConfigDirs, entries: names})
	}
}

func WithSecretKeysConfig(description string, names ...string) Option {
	return func(o *options) {
		o.configItems = append(o.configItems, configItem{description: description, configType: ConfigSecretKeys, entries: names})
	}
}

func WithInstanceName(name string) Option {
	return func(opts *options) {
		opts.instanceName = name
	}
}
