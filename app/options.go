package app

type options struct {
	withResources bool
}

type Option func(*options)

func WithResourcesEnabled(enabled bool) Option {
	return func(i *options) {
		i.withResources = enabled
	}
}
