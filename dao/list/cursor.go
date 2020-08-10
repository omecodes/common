package list

import (
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/utils/codec"
	"io"
)

type Cursor interface {
	HasNext() bool
	Next(o interface{}) (int64, error)
	io.Closer
}

type defaultCursor struct {
	inner    dao.Cursor
	encoding codec.Codec
}

func (d *defaultCursor) HasNext() bool {
	return d.inner.HasNext()
}

func (d *defaultCursor) Next(o interface{}) (int64, error) {
	i, err := d.inner.Next()
	if err != nil {
		return 0, err
	}

	row := i.(*Row)
	data := []byte(row.encoded)
	return row.index, d.encoding.Decode(data, o)
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
