package mapped

import "github.com/zoenion/common/dao"

func sqlScanRow(row dao.Row) (interface{}, error) {
	var r Row
	err := row.Scan(&r.first, &r.second, &r.encoded)
	return &r, err
}

func sqlScanPair(row dao.Row) (interface{}, error) {
	var p Pair
	err := row.Scan(&p.key, &p.encoded)
	return &p, err
}
