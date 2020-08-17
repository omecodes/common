package crypt

import (
	"io"
)

type AESGCMDecryptWrapper struct {
	key  []byte
	opts []ReadOption
}

func (d *AESGCMDecryptWrapper) Wrap(reader io.Reader) io.Reader {
	r := newAesGCMDecryptReader(d.key, reader, d.opts...)
	return r
}

func NewDecryptWrapper(key []byte, opts ...ReadOption) *AESGCMDecryptWrapper {
	return &AESGCMDecryptWrapper{
		key:  key,
		opts: opts,
	}
}
