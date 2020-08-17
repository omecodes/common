package crypt

import "io"

type ReaderWrapper interface {
	Wrap(reader io.Reader) io.Reader
}

type WriterWrapper interface {
	Wrap(writer io.Writer) io.Writer
}
