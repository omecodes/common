package app

type options struct {
	withResources bool
	configItems   []configItem
}

type Option func(*options)

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
