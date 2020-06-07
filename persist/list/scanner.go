package list

import "github.com/omecodes/common/dao"

func scanRow(row dao.Row) (interface{}, error) {
	var r Row
	err := row.Scan(&r.index, &r.encoded)
	return &r, err
}

func scanInt(row dao.Row) (interface{}, error) {
	var val int64
	return val, row.Scan(&val)
}
