package persist

import (
	"github.com/zoenion/common/dao"
)

type pair struct {
	Key   string
	Value []byte
}

func scanBytes(row dao.Row) (interface{}, error) {
	var bytes []byte
	return bytes, row.Scan(&bytes)
}
