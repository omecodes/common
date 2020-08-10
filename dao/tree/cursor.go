package tree

import (
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/utils/codec"
	"io"
)

type Cursor interface {
	HasNext() bool
	Next(o interface{}) (string, error)
	io.Closer
}

type defaultCursor struct {
	inner dao.Cursor
	codec codec.Codec
}

func (d *defaultCursor) HasNext() bool {
	return d.inner.HasNext()
}

func (d *defaultCursor) Next(o interface{}) (string, error) {
	i, err := d.inner.Next()
	if err != nil {
		return "", err
	}

	row := i.(*EncodedRow)
	data := []byte(row.encoded)
	return row.nodeName, d.codec.Decode(data, o)
}

func (d *defaultCursor) Close() error {
	return d.inner.Close()
}

func newCursor(cursor dao.Cursor, codec codec.Codec) Cursor {
	return &defaultCursor{
		inner: cursor,
		codec: codec,
	}
}
