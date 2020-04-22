package mapped

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/dao"
	"io"
)

type Cursor interface {
	HasNext() bool
	Next(o interface{}) error
	io.Closer
}

type defaultCursor struct {
	inner    dao.Cursor
	encoding codec.Codec
}

func (d *defaultCursor) HasNext() bool {
	return d.inner.HasNext()
}

func (d *defaultCursor) Next(o interface{}) error {
	i, err := d.inner.Next()
	if err != nil {
		return err
	}

	data := []byte(i.(string))
	return d.encoding.Decode(data, o)
}

func (d *defaultCursor) Close() error {
	return d.inner.Close()
}

func newCursor(cursor dao.Cursor, codec codec.Codec) Cursor {
	return &defaultCursor{
		inner:    cursor,
		encoding: codec,
	}
}

type MapCursor interface {
	HasNext() bool
	Next(o interface{}) (string, error)
	io.Closer
}

type defaultMapCursor struct {
	inner dao.Cursor
	codec codec.Codec
}

func (d *defaultMapCursor) HasNext() bool {
	return d.inner.HasNext()
}

func (d *defaultMapCursor) Next(o interface{}) (string, error) {
	i, err := d.inner.Next()
	if err != nil {
		return "", err
	}
	pair := i.(*Pair)

	data := []byte(pair.encoded)
	return pair.key, d.codec.Decode(data, o)
}

func (d *defaultMapCursor) Close() error {
	return d.inner.Close()
}

func newMapCursor(cursor dao.Cursor, codec codec.Codec) MapCursor {
	return &defaultMapCursor{
		inner: cursor,
		codec: codec,
	}
}
