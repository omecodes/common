package crypt

import (
	"bytes"
	"io"
)

type AESGCMEncryptWrapper struct {
	key        []byte
	options    []ReadOption
	outputSize int64
}

func (d *AESGCMEncryptWrapper) Wrap(reader io.Reader) io.Reader {
	return newAesGCMEncryptReader(d.key, reader, d.options...)
}

func (d *AESGCMEncryptWrapper) WithOutputSize(inputSize int64) int64 {
	opts := &readOptions{}
	for _, o := range d.options {
		o(opts)
	}
	if opts.blockSize == 0 {
		opts.blockSize = defaultBlockSize
	}

	plainSize := inputSize

	if opts.withRange {
		if opts.limit == -1 {
			plainSize = plainSize - opts.offset
		} else {
			plainSize = opts.limit - opts.offset
		}
	}

	var header header
	// header.Options = new(options)
	header.Nonce = make([]byte, 12)

	buffer := bytes.NewBuffer([]byte{})
	headerSize, _ := header.Write(buffer)

	blockCount := plainSize / opts.blockSize
	encryptedBlockSize := blockCount * (opts.blockSize + int64(headerSize) + aesGCMAuthTagSize)
	if blockCount*opts.blockSize != plainSize {
		lastBLockSize := (plainSize % opts.blockSize) + int64(headerSize) + aesGCMAuthTagSize
		encryptedBlockSize += lastBLockSize
	}
	return encryptedBlockSize
}

func NewEncryptWrapper(key []byte, opts ...ReadOption) *AESGCMEncryptWrapper {
	return &AESGCMEncryptWrapper{
		key:     key,
		options: opts,
	}
}
