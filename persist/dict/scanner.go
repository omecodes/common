package dict

import "github.com/zoenion/common/dao"

func scanRow(row dao.Row) (interface{}, error) {
	p := new(Row)
	return p, row.Scan(&p.key, &p.encoded)
}

func scanBool(row dao.Row) (interface{}, error) {
	var result int
	err := row.Scan(&result)
	return result == 1, err
}
