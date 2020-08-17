package crypt

type ReadOption func(*readOptions)

type readOptions struct {
	withRange bool
	blockSize int64
	offset    int64
	limit     int64
}

func WithOffset(offset int64) ReadOption {
	return func(opts *readOptions) {
		opts.offset = offset
		if !opts.withRange {
			opts.withRange = offset > 0
		}
	}
}

func WithLimit(limit int64) ReadOption {
	return func(opts *readOptions) {
		opts.limit = limit
		if !opts.withRange {
			opts.withRange = limit > 0
		}
	}
}

func WithBlockSize(size int64) ReadOption {
	return func(opts *readOptions) {
		opts.blockSize = size
	}
}
