package mapping

import (
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/utils/codec"
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

	row := i.(*Row)
	return d.encoding.Decode([]byte(row.encoded), o)
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
