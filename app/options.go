package app

import "github.com/zoenion/common/conf"

type options struct {
	startCMDFunc         func()
	afterConfigure       func(cfg conf.Map, outputFilename string) error
	version              string
	withResources        bool
	configItems          []configItem
	customAppDataDirPath string
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

func WithAfterConfigure(f func(cfg conf.Map, outputFilename string) error) Option {
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
