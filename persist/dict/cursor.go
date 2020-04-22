package dict

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/dao"
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

	r := i.(*Row)

	data := []byte(r.encoded)
	return r.key, d.codec.Decode(data, o)
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
