package tree

import "github.com/zoenion/common/dao"

func scanTreeRow(row dao.Row) (interface{}, error) {
	var tr TreeRow
	err := row.Scan(&tr.id, &tr.nodePath)
	return &tr, err
}

func scanEncodedRow(row dao.Row) (interface{}, error) {
	var er EncodedRow
	err := row.Scan(&er.parent, &er.nodeName, &er.encoded)
	return &er, err
}
